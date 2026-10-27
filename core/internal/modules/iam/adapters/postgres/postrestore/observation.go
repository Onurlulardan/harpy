package postrestore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Observation struct {
	State         string
	ReceiptSHA256 *string
	ReactivatedAt *time.Time
}

type Observer struct {
	within func(context.Context, func(database) error) error
}

func (observer *Observer) ObserveReactivation(
	ctx context.Context,
	environmentID uuid.UUID,
	restoreOperationID uuid.UUID,
) (Observation, error) {
	if observer == nil || observer.within == nil ||
		environmentID == uuid.Nil || restoreOperationID == uuid.Nil {
		return Observation{},
			fmt.Errorf("post-restore reactivation observation is unavailable")
	}
	var result Observation
	err := observer.within(
		ctx,
		func(database database) error {
			return database.QueryRow(
				ctx, observeReactivationSQL, environmentID, restoreOperationID,
			).Scan(
				&result.State, &result.ReceiptSHA256, &result.ReactivatedAt,
			)
		},
	)
	if err != nil {
		return Observation{}, err
	}
	if result.State != "pending" && result.State != "reactivated" {
		return Observation{},
			fmt.Errorf("post-restore reactivation observation differs")
	}
	if result.State == "pending" &&
		(result.ReceiptSHA256 != nil || result.ReactivatedAt != nil) {
		return Observation{},
			fmt.Errorf("pending post-restore observation carries a receipt")
	}
	if result.State == "reactivated" &&
		(result.ReceiptSHA256 == nil || result.ReactivatedAt == nil) {
		return Observation{},
			fmt.Errorf("reactivated post-restore observation lacks a receipt")
	}
	return result, nil
}
