package iam

import (
	"fmt"
	"time"
)

// NewWorkerSystemLayoutAccessRuntime constructs the layout access port for
// core-worker. Worker has no HTTP bearer surface, so Verify stays fail-closed.
func NewWorkerSystemLayoutAccessRuntime() (*SystemLayoutAccessRuntime, error) {
	authorizer, err := newSystemLayoutAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM worker system layout authorizer: %w", err)
	}
	return &SystemLayoutAccessRuntime{authorizer: authorizer, clock: time.Now}, nil
}
