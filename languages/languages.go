package languages

import (
	"github.com/jjs-dev/ci-config-gen/actions"
	"github.com/jjs-dev/ci-config-gen/config"
)

type JobSet struct {
	CI []actions.Job
}

type Language interface {
	Name() string
	Used(repoRoot string) bool
	Make(config config.CiConfig) JobSet
	MakeE2eCacheStep() (bool, actions.Step)
}

func MakeLanguages() []Language {
	return []Language{
		makeLanguageForGo(),
		makeLanguageForRust(),
	}
}
