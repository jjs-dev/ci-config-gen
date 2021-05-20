package languages

import (
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
	"github.com/jjs-dev/ci-config-gen/config"
)

type langGo struct{}

func (langGo) Name() string {
	return "golang"
}

func (langGo) Used(root string) bool {
	return checkPathExists(path.Join(root, "go.mod"))
}

func MakeSetupGoStep() actions.Step {
	return actions.Step{
		Name: "Install golang",
		Uses: "actions/setup-go@v2",
		With: map[string]string{
			"go-version": "1.16.4",
		},
	}
}

func (langGo) Make(_repoRoot string, config config.CiConfig) JobSet {
	return JobSet{
		CI: []actions.Job{
			{
				Name:    "go-lint",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					MakeSetupGoStep(),
					{
						Name: "Run linter",
						Uses: "golangci/golangci-lint-action@v2",
						With: map[string]string{
							"version":              "latest",
							"args":                 "--enable=gofmt",
							"skip-go-installation": "false",
						},
					},
				},
			},
			{
				Name:    "go-test",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					MakeSetupGoStep(),
					{
						Name: "Run tests",
						Run:  "go test .",
					},
				},
			},
		},
	}
}

func (langGo) MakeE2eCacheStep() (bool, actions.Step) {
	return false, actions.Step{}
}

func (langGo) WriteAdditionalFiles(repoRoot string) error {
	return nil
}

func makeLanguageForGo() Language {
	return langGo{}
}
