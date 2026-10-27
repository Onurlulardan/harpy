package platformlicense

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (service *Service) Read(
	ctx context.Context,
	environmentID uuid.UUID,
) (License, error) {
	if service == nil || service.store == nil || ctx == nil || ctx.Err() != nil ||
		!validEnvironment(environmentID) {
		return License{}, ErrInvalid
	}
	value, err := service.store.ReadLicense(ctx, environmentID)
	if err != nil {
		return License{}, fmt.Errorf("read platform license: %w", err)
	}
	if !validLicense(value, environmentID) {
		return License{}, ErrUnavailable
	}
	return value, nil
}
