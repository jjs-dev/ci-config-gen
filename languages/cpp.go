package languages

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/jjs-dev/ci-config-gen/actions"
	"github.com/jjs-dev/ci-config-gen/config"
)

type langCpp struct{}

func (langCpp) Name() string {
	return "cpp"
}

func findCmakeLists(root string) []string {
	g := filepath.Clean(root) + "/*/CMakeLists.txt"
	matches, err := filepath.Glob(g)
	if err != nil {
		log.Fatalf("invalid glob used: %v", err)
	}

	dirs := make([]string, 0)
	for _, cmakeList := range matches {
		dirName := filepath.Dir(cmakeList)
		dirName = strings.TrimPrefix(dirName, root)
		dirs = append(dirs, dirName)
	}

	return dirs
}

func (langCpp) Used(repoRoot string) bool {
	m := findCmakeLists(repoRoot)
	return len(m) > 0
}

func (langCpp) MakeE2eCacheStep() (bool, actions.Step) {
	return false, actions.Step{}
}

func (langCpp) Make(repoRoot string, config config.CiConfig) JobSet {
	m := findCmakeLists(repoRoot)
	fmt.Println(m)
	lintJob := actions.Job{
		Name:    "cpp-lint",
		RunsOn:  actions.UbuntuRunner,
		Timeout: config.JobTimeout,
		Steps: []actions.Step{
			actions.MakeCheckoutStep(),
			{
				Name: "Install dependencies",
				Run:  "sudo apt-get install -y clang-tools",
			},
			{
				Name: "Prepare report directory",
				Run:  "mkdir analyzer-report",
			},
		},
	}
	for _, dir := range m {
		stepConfigure := actions.Step{
			Name: fmt.Sprintf("Configure %s", dir),
			Run:  fmt.Sprintf("cmake -S %s -B %s/cmake-build -DCMAKE_EXPORT_COMPILE_COMMANDS=On", dir, dir),
		}
		stepLint := actions.Step{
			Name: fmt.Sprintf("Lint %s", dir),
			Run:  fmt.Sprintf("scan-build -o analyzer-report make -C %s/cmake-build -j4", dir),
		}
		lintJob.Steps = append(lintJob.Steps, stepConfigure, stepLint)
	}
	stepCheckNoErrors := actions.Step{
		Name: "Check that report is empty",
		Run:  "[ -z \"$(ls -A analyzer-report)\" ]",
	}
	// TODO: upload report
	lintJob.Steps = append(lintJob.Steps, stepCheckNoErrors)
	return JobSet{
		CI: []actions.Job{lintJob},
	}
}

func makeLanguageForCpp() Language {
	return langCpp{}
}
