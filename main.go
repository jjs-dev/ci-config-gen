package main

import (
	"flag"
	"fmt"
	"github.com/jjs-dev/ci-config-gen/actions"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
)

func validateWorkflow(workflow actions.Workflow) {
	for _, job := range workflow.Jobs {
		if job.RunsOn == "" {
			log.Fatalf("In workflow %s a job has missing runs-on", workflow.Name)
		}
	}
}

func preprocessWorkflow(workflow actions.Workflow) actions.Workflow {
	validateWorkflow(workflow)
	// no modifications currently
	return workflow
}

func writeWorkflow(out string, workflow actions.Workflow) {

	y, err := yaml.Marshal(preprocessWorkflow(workflow))
	if err != nil {
		log.Fatalf("failed to serialize workflow %v", err)
	}

	workflowPath :=
		path.Join(out, ".github/workflows", fmt.Sprintf("%s.yaml", workflow.Name))
	err = os.WriteFile(workflowPath, y, 0o755)
	if err != nil {
		log.Fatalf("failed to write workflow %s: %v", workflowPath, err)
	}
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

	metaWorkflow := makeMetaWorkflow()
	writeWorkflow(*out, metaWorkflow)

	perLanguageJobs := make([]JobSet, 0)
	if hasRust(*repoRoot) {
		log.Println("Generating rust jobs")
		perLanguageJobs = append(perLanguageJobs, makeRustJobs())
	}
	if hasGo(*repoRoot) {
		log.Println("Generating golang jobs")
		perLanguageJobs = append(perLanguageJobs, makeGoJobs())
	}

	ciWorkflow := makeCiWorkflow(perLanguageJobs)
	writeWorkflow(*out, ciWorkflow)

	if !config.NoPublish {
		log.Println("Generating publish workflow")
		publishWorkflow := makePublishWorkflow(*repoRoot)
		writeWorkflow(*out, publishWorkflow)
	}
}

func makeMetaWorkflow() actions.Workflow {
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
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
					makeSetupGoStep(),
					{
						Run:  "go install -v github.com/jjs-dev/ci-config-gen@HEAD",
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

func makeCiWorkflow(jobsets []JobSet) actions.Workflow {
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
				RunsOn: actions.UbuntuRunner,
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

	for _, js := range jobsets {
		for _, job := range js.ci {
			w.Jobs[job.Name] = job
		}
	}

	return w
}

func makePublishWorkflow(root string) actions.Workflow {

	publishJob := actions.Job{
		RunsOn: actions.UbuntuRunner,
	}

	w := actions.Workflow{
		Name: "publish",
		On: actions.Trigger{
			PullRequest: actions.EmptyStruct{},
			Push: actions.PushTrigger{
				Branches: []string{"staging", "trying", "master", "force-publish"},
			},
		},
		Jobs: map[string]actions.Job{
			"publish": publishJob,
		},
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
