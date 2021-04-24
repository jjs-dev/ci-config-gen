package main

import (
	"github.com/jjs-dev/ci-config-gen/bors"
	"gotest.tools/v3/assert"
	"testing"
)

func TestMetaWorkflowValid(t *testing.T) {
	meta := makeMetaWorkflow(&bors.BorsConfig{}, false)
	err := meta.Validate()
	assert.NilError(t, err)
}
