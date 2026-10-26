package licenseauthority

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RestoreBinding struct {
	Binding           Binding
	Issuer            string
	Sequence          int64
	SHA256            string
	AuthorityRevision int64
	DatabaseNow       time.Time
}

type RestoreBindingStore interface {
	ResolveRestoreBinding(context.Context, uuid.UUID) (RestoreBinding, error)
}

func ResolveRestoreBinding(
	ctx context.Context,
	store RestoreBindingStore,
	environmentID uuid.UUID,
) (RestoreBinding, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || environmentID == uuid.Nil {
		return RestoreBinding{}, ErrAuthorityState
	}
	binding, err := store.ResolveRestoreBinding(ctx, environmentID)
	if err != nil {
		return RestoreBinding{}, err
	}
	if binding.Binding.EnvironmentID != environmentID ||
		!validBinding(binding.Binding) || !validText(binding.Issuer, 256) ||
		binding.Sequence < 1 || !digestPattern.MatchString(binding.SHA256) ||
		binding.AuthorityRevision < 1 || binding.DatabaseNow.IsZero() ||
		binding.DatabaseNow.Location() != time.UTC {
		return RestoreBinding{}, ErrAuthorityState
	}
	return binding, nil
}
