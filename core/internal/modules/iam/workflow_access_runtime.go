package iam

import "fmt"

type WorkflowAccessRuntime struct{ shared *FormsRecordAccessRuntime }

func NewWorkflowAccessRuntime(
	shared *FormsRecordAccessRuntime,
) (*WorkflowAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Workflow access dependencies are required")
	}
	return &WorkflowAccessRuntime{shared: shared}, nil
}
