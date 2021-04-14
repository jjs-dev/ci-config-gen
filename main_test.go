package main

import "testing"

func TestMetaWorkflowValid(t *testing.T) {
	meta := makeMetaWorkflow()
	validateWorkflow(meta)
}
