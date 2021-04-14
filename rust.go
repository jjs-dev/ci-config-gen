package main

import (
	"github.com/jjs-dev/ci-config-gen/actions"
	"path"
)

/*
type rustConfig struct {

}*/

func hasRust(root string) bool {
	return checkPathExists(path.Join(root, "Cargo.toml"))
}

func makeRustJobs() JobSet {
	return JobSet{
		ci: []actions.Job{
			{
				Name:   "rustfmt",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
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
		},
	}
}
