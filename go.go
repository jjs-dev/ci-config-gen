package main

import (
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
)

func hasGo(root string) bool {
	return checkPathExists(path.Join(root, "go.mod"))
}

func makeGoJobs() JobSet {
	return JobSet{
		ci: []actions.Job{
			{
				Name:   "lint",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
					{
						Name: "Run linter",
						Uses: "golangci/golangci-lint-action@v2",
						With: map[string]string{
							"version": "latest",
							"args":    "--enable=gofmt",
						},
					},
				},
			},
			{
				Name:   "test",
				RunsOn: actions.UbuntuRunner,
				Steps: []actions.Step{
					makeCheckoutStep(),
					{
						Name: "Run tests",
						Run:  "go test .",
					},
				},
			},
		},
	}
}
