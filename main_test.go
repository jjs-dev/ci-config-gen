package main

import (
	"testing"

	"github.com/jjs-dev/ci-config-gen/bors"
)

func TestMetaWorkflowValid(t *testing.T) {
	meta := makeMetaWorkflow(&bors.BorsConfig{})
	meta.Validate()
}
