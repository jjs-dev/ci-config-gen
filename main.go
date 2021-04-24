package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
	"github.com/jjs-dev/ci-config-gen/bors"
	"gopkg.in/yaml.v2"
)

func preprocessWorkflow(workflow actions.Workflow) actions.Workflow {
	err := workflow.Validate()
	if err != nil {
		log.Fatalf("Workflow %s in invalid: %v", workflow.Name, err)
	}
	// no modifications currently
	return workflow
}

func emitFile(out string, relName string, data []byte) {
	fullPath := path.Join(out, relName)
	err := os.WriteFile(fullPath, data, 0o755)
	if err != nil {
		log.Fatalf("failed to write %s: %v", relName, err)
	}
}

func writeWorkflow(out string, workflow actions.Workflow) {
	alert := "# GENERATED FILE DO NOT EDIT\n"
	y, err := yaml.Marshal(preprocessWorkflow(workflow))
	data := append([]byte(alert), y...)
	if err != nil {
		log.Fatalf("failed to serialize workflow %v", err)
	}
	emitFile(out, fmt.Sprintf(".github/workflows/%s.yaml", workflow.Name), data)
}

func checkPathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func main() {
	repoRoot := flag.String("repo-root", "", "path to root directory of the repository to generate config for")
	out := flag.String("output", "", "directory which will contain generated workflow files. defaults to $(repo-root)")

	flag.Parse()
	if *repoRoot == "" {
		log.Fatal("--repo-root not provided")
	}

	if *out == "" {
		*out = *repoRoot
	}

	config, err := loadCiConfig(*repoRoot)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("loaded config: %+v", config)

	borsConfig := &bors.BorsConfig{}
	borsConfig.ApplyDefaults()
	borsConfig.Timeout = config.BuildTimeout * 60

	metaWorkflow := makeMetaWorkflow(borsConfig, config.InternalHackForGenerator)
	writeWorkflow(*out, metaWorkflow)

	perLanguageJobs := make([]JobSet, 0)
	if hasRust(*repoRoot) {
		log.Println("Generating rust jobs")
		perLanguageJobs = append(perLanguageJobs, makeRustJobs(config))
	}
	if hasGo(*repoRoot) {
		log.Println("Generating golang jobs")
		perLanguageJobs = append(perLanguageJobs, makeGoJobs(config))
	}

	ciWorkflow := makeCiWorkflow(perLanguageJobs, config, *repoRoot, borsConfig)
	writeWorkflow(*out, ciWorkflow)

	if !config.NoPublish {
		log.Println("Generating publish workflow")
		publishWorkflow := makePublishWorkflow(*repoRoot, config)
		writeWorkflow(*out, publishWorkflow)
		script := generatePublishImageScript(config)
		emitFile(*out, "ci/publish-images.sh", []byte(script))
	}
	log.Println("Generating bors config")
	borsConfigBytes, err := borsConfig.Serialize()
	if err != nil {
		log.Fatal(err)
	}
	emitFile(*out, "bors.toml", borsConfigBytes)
}

func makeMetaWorkflow(bc *bors.BorsConfig, useLocal bool) actions.Workflow {
	bc.AddJob("check-ci-config")

	var genLocation string
	if useLocal {
		genLocation = "."
	} else {
		genLocation = "github.com/jjs-dev/ci-config-gen@HEAD"
	}

	return actions.Workflow{
		Name: "meta",
		On: actions.Trigger{
			PullRequest: actions.EmptyStruct{},
			Push: actions.PushTrigger{
				Branches: []string{"staging", "trying", "master"},
			},
		},
		Jobs: map[string]actions.Job{
			"check-ci-config": {
				RunsOn:  actions.UbuntuRunner,
				Timeout: 1,
				Steps: []actions.Step{
					makeCheckoutStep(),
					makeSetupGoStep(),
					{
						Run:  fmt.Sprintf("go install -v %s", genLocation),
						Name: "Install ci-config-gen",
					},
					{
						Run:  "ci-config-gen --repo-root .",
						Name: "Run co-config-gen",
					},
					{
						Name: "Verify CI configuration is up-to-date",
						Run:  "git diff --exit-code",
					},
				},
			},
		},
	}

}

func makeCiE2eJob(root string, config ciConfig) (actions.Job, actions.Job) {
	buildSteps := []actions.Step{
		makeCheckoutStep(),
	}
	if hasRust(root) {
		buildSteps = append(buildSteps, makeRustCacheStep())
	}
	buildSteps = append(buildSteps, actions.Step{
		Name: "Build e2e artifacts",
		Run:  "bash ci/e2e-build.sh",
	}, actions.Step{
		Name: "Upload e2e artifacts",
		Uses: "actions/upload-artifact@v2",
		With: map[string]string{
			"name":           "e2e-artifacts",
			"path":           "e2e-artifacts",
			"retention-days": "2",
		},
	})

	build := actions.Job{
		RunsOn:  actions.UbuntuRunner,
		Steps:   buildSteps,
		Timeout: config.JobTimeout,
		Env: map[string]string{
			"DOCKER_BUILDKIT": "1",
		},
	}
	run := actions.Job{
		RunsOn:  actions.UbuntuRunner,
		Needs:   "e2e-build",
		Timeout: config.JobTimeout,
		Steps: []actions.Step{
			makeCheckoutStep(),
			{
				Name: "Download e2e artifacts",
				Uses: "actions/download-artifact@v2",
				With: map[string]string{
					"name": "e2e-artifacts",
					"path": "e2e-artifacts",
				},
			},
			{
				Name: "Execute tests",
				Run:  "bash ci/e2e-run.sh",
			},
			{
				Name: "Upload logs",
				Uses: "actions/upload-artifact@v2",
				With: map[string]string{
					"name":           "e2e-logs",
					"path":           "e2e-logs",
					"retention-days": "2",
				},
			},
		},
	}

	return build, run
}

func makeCiWorkflow(jobsets []JobSet, config ciConfig, repoRoot string, bc *bors.BorsConfig) actions.Workflow {
	w := actions.Workflow{
		Name: "ci",
		On: actions.Trigger{
			PullRequest: actions.EmptyStruct{},
			Push: actions.PushTrigger{
				Branches: []string{"staging", "trying", "master"},
			},
		},
		Jobs: map[string]actions.Job{
			"misspell": {
				RunsOn:  actions.UbuntuRunner,
				Timeout: 2,
				Steps: []actions.Step{
					makeCheckoutStep(),
					{
						Name: "run spellcheck",
						Uses: "reviewdog/action-misspell@v1",
						With: map[string]string{
							"github_token": "${{ secrets.GITHUB_TOKEN }}",
							"locale":       "US",
						},
					},
				},
			},
		},
	}

	if !config.NoE2e {
		e2eBuild, e2eRun := makeCiE2eJob(repoRoot, config)
		bc.AddJob("e2e-build")
		bc.AddJob("e2e-run")
		w.Jobs["e2e-build"] = e2eBuild
		w.Jobs["e2e-run"] = e2eRun
	}

	for _, js := range jobsets {
		for _, job := range js.ci {
			bc.AddJob(job.Name)
			w.Jobs[job.Name] = job
		}
	}

	return w
}

func makeCheckoutStep() actions.Step {
	return actions.Step{
		Name: "Fetch sources",
		Uses: "actions/checkout@v2",
	}
}

type JobSet struct {
	ci []actions.Job
}
