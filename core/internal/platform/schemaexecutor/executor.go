package schemaexecutor

import (
	"fmt"
	"time"
)

const (
	LeaseDuration            = 60 * time.Second
	HeartbeatInterval        = 20 * time.Second
	DatabaseOperationTimeout = HeartbeatInterval / 2
	IdlePollInterval         = 500 * time.Millisecond
	MaximumAttempts          = 8
)

type Executor struct {
	persistence       Persistence
	executorID        string
	newLeaseID        func() (string, error)
	heartbeatInterval time.Duration
}

type Dependencies struct {
	Persistence Persistence
	ExecutorID  string
	NewLeaseID  func() (string, error)
}

func New(dependencies Dependencies) (*Executor, error) {
	if dependencies.Persistence == nil || dependencies.NewLeaseID == nil {
		return nil, fmt.Errorf("schema executor dependencies are incomplete")
	}
	if _, err := canonicalUUID(dependencies.ExecutorID); err != nil {
		return nil, fmt.Errorf("schema executor identity is invalid")
	}
	return &Executor{
		persistence:       dependencies.Persistence,
		executorID:        dependencies.ExecutorID,
		newLeaseID:        dependencies.NewLeaseID,
		heartbeatInterval: HeartbeatInterval,
	}, nil
}
