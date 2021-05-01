package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
	"github.com/jjs-dev/ci-config-gen/bors"
	"github.com/jjs-dev/ci-config-gen/config"
	"github.com/jjs-dev/ci-config-gen/languages"
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

	config, err := config.Load(*repoRoot)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("loaded config: %+v", config)

	borsConfig := &bors.BorsConfig{}
	borsConfig.ApplyDefaults()
	borsConfig.Timeout = config.BuildTimeout * 60

	metaWorkflow := makeMetaWorkflow(borsConfig, config.InternalHackForGenerator)
	writeWorkflow(*out, metaWorkflow)

	langs := languages.MakeLanguages()

	ciWorkflow := makeCiWorkflow(langs, config, *repoRoot, borsConfig)
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

	var fetchGenerator actions.Step
	var generatorLocation string
	if useLocal {
		fetchGenerator = actions.Step{
			Name: "No-op",
			Run:  "echo OK",
		}
		generatorLocation = "."
	} else {
		fetchGenerator = actions.Step{
			Name: "Fetch generator sources",
			Run:  "git clone https://github.com/jjs-dev/ci-config-gen ./gen",
		}
		generatorLocation = "./gen"
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
					actions.MakeCheckoutStep(),
					languages.MakeSetupGoStep(),
					fetchGenerator,
					{
						Run:  fmt.Sprintf("cd %s && go install -v .", generatorLocation),
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

func makeCiE2eJob(root string, config config.CiConfig, languages []languages.Language) (actions.Job, actions.Job) {
	buildSteps := []actions.Step{
		actions.MakeCheckoutStep(),
	}
	for _, lang := range languages {
		if lang.Used(root) {
			needsCache, cacheStep := lang.MakeE2eCacheStep()
			if !needsCache {
				continue
			}
			buildSteps = append(buildSteps, cacheStep)
		}
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
			actions.MakeCheckoutStep(),
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

func makeCiWorkflow(langs []languages.Language, config config.CiConfig, repoRoot string, bc *bors.BorsConfig) actions.Workflow {
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
					actions.MakeCheckoutStep(),
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

	perLanguageJobs := make([]languages.JobSet, 0)

	for _, lang := range langs {
		if lang.Used(repoRoot) {
			log.Printf("Generating %s CI jobs", lang.Name())
			perLanguageJobs = append(perLanguageJobs, lang.Make(repoRoot, config))
		}
	}

	if !config.NoE2e {
		e2eBuild, e2eRun := makeCiE2eJob(repoRoot, config, langs)
		bc.AddJob("e2e-build")
		bc.AddJob("e2e-run")
		w.Jobs["e2e-build"] = e2eBuild
		w.Jobs["e2e-run"] = e2eRun
	}

	for _, js := range perLanguageJobs {
		for _, job := range js.CI {
			bc.AddJob(job.Name)
			w.Jobs[job.Name] = job
		}
	}

	return w
}
