package actions

const (
	UbuntuRunner = "ubuntu-20.04"
)

type Workflow struct {
	Name string
	On   Trigger
	Jobs map[string]Job
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
	Name   string            `yaml:",omitempty"`
	If     string            `yaml:",omitempty"`
	Needs  string            `yaml:",omitempty"`
	Env    map[string]string `yaml:",omitempty"`
	RunsOn string            `yaml:"runs-on"`
	Steps  []Step
}

type Step struct {
	Id   string            `yaml:",omitempty"`
	Name string            `yaml:",omitempty"`
	If   string            `yaml:",omitempty"`
	Uses string            `yaml:",omitempty"`
	Run  string            `yaml:",omitempty"`
	With map[string]string `yaml:",omitempty"`
}
