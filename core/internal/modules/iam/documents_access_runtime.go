package iam

import "fmt"

type DocumentsAccessRuntime struct{ shared *FormsRecordAccessRuntime }

func NewDocumentsAccessRuntime(shared *FormsRecordAccessRuntime) (*DocumentsAccessRuntime, error) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Documents access dependencies are required")
	}
	return &DocumentsAccessRuntime{shared: shared}, nil
}
