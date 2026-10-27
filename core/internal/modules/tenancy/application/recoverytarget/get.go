package recoverytarget

import (
	"context"
	"errors"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

var ErrUnavailable = errors.New("tenancy recovery target is unavailable")

type Reader interface {
	Get(context.Context, string) (tenantapi.RecoveryTarget, bool, error)
}

type Service struct {
	reader  Reader
	purpose string
}

func New(reader Reader, purpose string) (*Service, error) {
	if reader == nil || purpose != "backup" && purpose != "restore" {
		return nil, ErrUnavailable
	}
	return &Service{reader: reader, purpose: purpose}, nil
}

func (service *Service) Get(ctx context.Context) (tenantapi.RecoveryTarget, error) {
	if service == nil || ctx == nil || ctx.Err() != nil {
		return tenantapi.RecoveryTarget{}, ErrUnavailable
	}
	target, found, err := service.reader.Get(ctx, service.purpose)
	if err != nil || !found {
		return tenantapi.RecoveryTarget{}, ErrUnavailable
	}
	return target, nil
}
