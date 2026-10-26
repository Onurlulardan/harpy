package ports

import (
	"context"
	"time"
)

type RestorePlanSigningRequest struct {
	EnvironmentID, OperationID string
	Sequence                   int64
	Payload                    []byte
	IssuedAt, ExpiresAt        time.Time
}

type RestorePlanSigningResult struct {
	Generation int64
	Document   []byte
}

type RestorePlanSigner interface {
	SignRestorePlan(context.Context, RestorePlanSigningRequest) (RestorePlanSigningResult, error)
}
