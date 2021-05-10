package main

import (
	"testing"

	"github.com/jjs-dev/ci-config-gen/bors"
	"github.com/jjs-dev/ci-config-gen/config"
	"gotest.tools/v3/assert"
)

func TestMetaWorkflowValid(t *testing.T) {
	cfg := config.CiConfig{
		Codegen:    true,
		JobTimeout: 1,
	}
	meta := makeMetaWorkflow(&bors.BorsConfig{}, cfg)
	err := meta.Validate()
	assert.NilError(t, err)
}
