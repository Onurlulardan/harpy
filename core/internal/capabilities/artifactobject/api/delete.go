package api

import (
	"context"
	"time"
)

type DeleteResult struct {
	ProviderRequestID    string
	DeleteResponseSHA256 [32]byte
	AbsenceProbeSHA256   [32]byte
	VerifiedAbsentAt     time.Time
}

type Deleter interface {
	DeleteExact(context.Context, Identity) (DeleteResult, error)
}
