package actions

import "fmt"

const (
	UbuntuRunner = "ubuntu-20.04"
)

type Workflow struct {
	Name string
	On   Trigger
	Jobs map[string]Job
}

func (w Workflow) Validate() error {
	for jobName, job := range w.Jobs {
		jobErr := job.Validate()
		if job.Name != "" && job.Name != jobName {
			jobErr = fmt.Errorf("job name mismatch: named in map as %s, but name is %s", jobName, job.Name)
		}
		if jobErr != nil {
			return fmt.Errorf("invalid job %s: %w", jobName, jobErr)
		}
	}
	return nil
}

type Trigger struct {
	PullRequest EmptyStruct `yaml:"pull_request"`
	Push        PushTrigger
}

type PushTrigger struct {
	Branches []string
}

type EmptyStruct struct{}

type Job struct {
	Name    string            `yaml:",omitempty"`
	If      string            `yaml:",omitempty"`
	Needs   string            `yaml:",omitempty"`
	Env     map[string]string `yaml:",omitempty"`
	RunsOn  string            `yaml:"runs-on"`
	Timeout int               `yaml:"timeout-minutes"`
	Steps   []Step            `yaml:"steps"`
}

func (j Job) Validate() error {
	if j.RunsOn == "" {
		return fmt.Errorf("missing runs-on")
	}
	if j.Timeout == 0 {
		return fmt.Errorf("missing timeout-minutes")
	}
	return nil
}

type Step struct {
	Id   string            `yaml:",omitempty"`
	Name string            `yaml:",omitempty"`
	If   string            `yaml:",omitempty"`
	Uses string            `yaml:",omitempty"`
	Run  string            `yaml:",omitempty"`
	With map[string]string `yaml:",omitempty"`
}

func MakeCheckoutStep() Step {
	return Step{
		Name: "Fetch sources",
		Uses: "actions/checkout@v2",
	}
}
