package api

import (
	"context"
	"time"
)

type PostRestoreAwait struct {
	EnvironmentID      string
	RestoreOperationID string
}

type PostRestoreObservation struct {
	State         string
	ReceiptSHA256 *string
	ReactivatedAt *time.Time
}

type PostRestoreObserver interface {
	ObserveReactivation(
		context.Context, string, string,
	) (PostRestoreObservation, error)
}
