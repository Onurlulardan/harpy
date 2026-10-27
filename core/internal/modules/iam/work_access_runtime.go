package iam

import "fmt"

type WorkAccessRuntime struct {
	shared *FormsRecordAccessRuntime
}

func NewWorkAccessRuntime(
	shared *FormsRecordAccessRuntime,
) (*WorkAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Work access dependencies are required")
	}
	return &WorkAccessRuntime{shared: shared}, nil
}
