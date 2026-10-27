package iam

import "fmt"

type WorkAccessRuntime struct {
	shared *accessRequestDependencies
}

func NewWorkAccessRuntime(
	shared *accessRequestDependencies,
) (*WorkAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Work access dependencies are required")
	}
	return &WorkAccessRuntime{shared: shared}, nil
}
