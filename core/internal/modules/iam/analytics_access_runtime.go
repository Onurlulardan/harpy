package iam

import "fmt"

type AnalyticsAccessRuntime struct{ shared *FormsRecordAccessRuntime }

func NewAnalyticsAccessRuntime(
	shared *FormsRecordAccessRuntime,
) (*AnalyticsAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Analytics access dependencies are required")
	}
	return &AnalyticsAccessRuntime{shared: shared}, nil
}
