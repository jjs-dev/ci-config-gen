package main

import (
	"path"

	"github.com/jjs-dev/ci-config-gen/actions"
)

func hasGo(root string) bool {
	return checkPathExists(path.Join(root, "go.mod"))
}

func makeSetupGoStep() actions.Step {
	return actions.Step{
		Name: "Install golang",
		Uses: "actions/setup-go@v2",
		With: map[string]string{
			"go-version": "1.16.3",
		},
	}
}

func makeGoJobs(config ciConfig) JobSet {
	return JobSet{
		ci: []actions.Job{
			{
				Name:    "go-lint",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					makeCheckoutStep(),
					makeSetupGoStep(),
					{
						Name: "Run linter",
						Uses: "golangci/golangci-lint-action@v2",
						With: map[string]string{
							"version":              "latest",
							"args":                 "--enable=gofmt",
							"skip-go-installation": "false",
						},
					},
				},
			},
			{
				Name:    "go-test",
				RunsOn:  actions.UbuntuRunner,
				Timeout: config.JobTimeout,
				Steps: []actions.Step{
					makeCheckoutStep(),
					makeSetupGoStep(),
					{
						Name: "Run tests",
						Run:  "go test .",
					},
				},
			},
		},
	}
}
