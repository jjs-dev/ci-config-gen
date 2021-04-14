package main

import (
	"fmt"
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
)

const (
	CargoUdepsVersion = "0.1.20"
)

/*
type rustConfig struct {

}*/

func hasRust(root string) bool {
	return checkPathExists(path.Join(root, "Cargo.toml"))
}

func makeRustCacheStep() actions.Step {
	return actions.Step{
		Name: "Setup cache",
		Uses: "Swatinem/rust-cache@v1",
	}
}

func makeInstallTooclhainStep(channel string) actions.Step {
	/*name: Install Rust toolchain
	uses: actions-rs/toolchain@v1
	with:
	  toolchain: nightly-2020-08-28
	  components: clippy,rustfmt
	  override: true*/
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

func makeRustJobs() JobSet {

	compileCargoUdeps := `
cargo install cargo-udeps --locked --version %s
mkdir -p ~/udeps
cp $( which cargo-udeps ) ~/udeps`

	runCargoUdeps := `
export PATH=~/udeps:$PATH  
cargo udeps 
`

	return JobSet{
		ci: []actions.Job{
			{
				Name:   "rustfmt",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
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
				Name:   "rust-unit-tests",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
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
				Name:   "rust-unused-deps",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
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
				Name:   "rust-cargo-deny",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
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
				Name: "rust-lint",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
					{
						Name: "Run clippy",
						Uses: "actions-rs/cargo@v1",
						With: map[string]string{
							"command": "clippy",
							"args": "--workspace -- -Dwarnings",
						},
					},
				},
			},
		},
	}
}
