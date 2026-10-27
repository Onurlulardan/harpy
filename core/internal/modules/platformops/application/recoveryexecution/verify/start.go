package verify

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

var ErrTransition = errors.New("recovery verify transition failed")

type Store interface {
	Start(context.Context, ports.RecoveryExecutionIdentity) (bool, error)
}

type Service struct{ store Store }

func New(store Store) *Service {
	if store == nil {
		return nil
	}
	return &Service{store: store}
}

func (service *Service) Start(
	ctx context.Context,
	identity ports.RecoveryExecutionIdentity,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !identity.Valid() {
		return ErrTransition
	}
	changed, err := service.store.Start(ctx, identity)
	if err != nil || !changed {
		return ErrTransition
	}
	return nil
}
