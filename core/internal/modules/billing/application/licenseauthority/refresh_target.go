package licenseauthority

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RefreshTarget struct {
	Found       bool
	Ready       bool
	OperationID uuid.UUID
	Binding     Binding
	Revision    int64
	RetryAt     time.Time
	DatabaseNow time.Time
}

type RefreshTargetStore interface {
	AcquireRefreshTarget(context.Context) (RefreshTarget, error)
}

func AcquireRefreshTarget(
	ctx context.Context,
	store RefreshTargetStore,
) (RefreshTarget, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return RefreshTarget{}, ErrAuthorityState
	}
	target, err := store.AcquireRefreshTarget(ctx)
	if err != nil {
		return RefreshTarget{}, err
	}
	if !validRefreshTarget(target) {
		return RefreshTarget{}, ErrAuthorityState
	}
	return target, nil
}

func validRefreshTarget(target RefreshTarget) bool {
	if target.DatabaseNow.IsZero() || target.DatabaseNow.Location() != time.UTC ||
		target.RetryAt.IsZero() || target.RetryAt.Location() != time.UTC ||
		target.RetryAt.Before(target.DatabaseNow) {
		return false
	}
	if !target.Found {
		return !target.Ready && target.OperationID == uuid.Nil &&
			target.Binding == (Binding{}) && target.Revision == 0
	}
	return target.OperationID.Version() == uuid.Version(7) &&
		validBinding(target.Binding) && target.Revision > 0
}
