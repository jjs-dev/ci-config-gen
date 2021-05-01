package languages

import (
	"fmt"
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
	"github.com/jjs-dev/ci-config-gen/config"
)

const (
	CargoUdepsVersion = "0.1.20"
)

/*
type rustConfig struct {

}*/

type langRust struct{}

func (langRust) Name() string {
	return "rust"
}

func (langRust) Used(root string) bool {
	return checkPathExists(path.Join(root, "Cargo.toml"))
}

func makeRustCacheStep() actions.Step {
	return actions.Step{
		Name: "Setup cache",
		Uses: "Swatinem/rust-cache@v1",
	}
}

func (langRust) MakeE2eCacheStep() (bool, actions.Step) {
	return true, makeRustCacheStep()
}

func makeInstallTooclhainStep(channel string) actions.Step {
	return actions.Step{
		Name: fmt.Sprintf("Install %s toolchain", channel),
		Uses: "actions-rs/toolchain@v1",
		With: map[string]string{
			"toolchain":  channel,
			"components": "clippy,rustfmt",
			"override":   "true",
		},
	}
}

func (langRust) Make(_repoRoot string, config config.CiConfig) JobSet {

	compileCargoUdeps := `
cargo install cargo-udeps --locked --version %s
mkdir -p ~/udeps
cp $( which cargo-udeps ) ~/udeps`

	runCargoUdeps := `
export PATH=~/udeps:$PATH  
cargo udeps 
`

	return JobSet{
		CI: []actions.Job{
			{
				Name:    "rustfmt",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					makeInstallTooclhainStep("nightly"),
					{
						Name: "Check formatting",
						Uses: "actions-rs/cargo@v1",
						With: map[string]string{
							"command": "fmt",
							"args":    "-- --check",
						},
					},
				},
			},
			{
				Name:    "rust-unit-tests",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					makeRustCacheStep(),
					{
						Name: "Run unit tests",
						Uses: "actions-rs/cargo@v1",
						With: map[string]string{
							"command": "test",
						},
					},
				},
			},
			{
				Name:    "rust-unused-deps",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					makeInstallTooclhainStep("nightly"),
					makeRustCacheStep(),
					{
						Name: "Fetch prebuilt cargo-udeps",
						Id:   "cargo_udeps",
						Uses: "actions/cache@v2",
						With: map[string]string{
							"path": "~/udeps",
							"key":  fmt.Sprintf("udeps-bin-${{ runner.os }}-v%s", CargoUdepsVersion),
						},
					},
					{
						Name: "Install cargo-udeps",
						If:   "steps.cache_udeps.outputs.cache-hit != 'true'",
						Run:  fmt.Sprintf(compileCargoUdeps, CargoUdepsVersion),
					},
					{
						Name: "Run cargo-udeps",
						Run:  runCargoUdeps,
					},
				},
			},
			{
				Name:    "rust-cargo-deny",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					{
						Name: "Run cargo-deny",
						Uses: "EmbarkStudios/cargo-deny-action@v1",
						With: map[string]string{
							"command": "check all",
						},
					},
				},
			},
			{
				Name:    "rust-lint",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					actions.MakeCheckoutStep(),
					{
						Name: "Run clippy",
						Uses: "actions-rs/cargo@v1",
						With: map[string]string{
							"command": "clippy",
							"args":    "--workspace -- -Dwarnings",
						},
					},
				},
			},
		},
	}
}

func makeLanguageForRust() Language {
	return langRust{}
}
