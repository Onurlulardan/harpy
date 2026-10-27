package iam

import "fmt"

type FormsLifecycleAccessRuntime struct{ shared *FormsRecordAccessRuntime }

func NewFormsLifecycleAccessRuntime(
	shared *FormsRecordAccessRuntime,
) (*FormsLifecycleAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Forms lifecycle access dependencies are required")
	}
	return &FormsLifecycleAccessRuntime{shared: shared}, nil
}
