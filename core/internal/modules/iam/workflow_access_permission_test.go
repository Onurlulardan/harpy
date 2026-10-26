package iam

import "testing"

func TestWorkflowAccessPermissionsMapOnlyExactOperations(t *testing.T) {
	for key, target := range workflowAccessTargets {
		if target.resource == "" || target.action == "" {
			t.Fatalf("Workflow access target %q is incomplete", key)
		}
	}
	if len(workflowAccessTargets) != 20 {
		t.Fatalf("Workflow access target count = %d", len(workflowAccessTargets))
	}
	for _, rejected := range [][2]string{
		{"workflow.definition.read", "workflow.definition.create"},
		{"workflow.instance.read", "workflow.runtime.signal"},
		{"workflow.task.read", "workflow.human_task.decide"},
		{"workflow.publication.publish", "workflow.runtime.start"},
	} {
		if _, found := resolveWorkflowAccess(rejected[0], rejected[1]); found {
			t.Fatalf("unexpected Workflow access target: %v", rejected)
		}
	}
}
