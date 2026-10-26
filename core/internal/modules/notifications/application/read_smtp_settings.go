package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *SMTPSettingsApplication) ReadSMTPSettings(ctx context.Context,
	scope api.Scope,
) (api.SMTPSettingsProjection, error) {
	if application == nil || ctx == nil || ctx.Err() != nil {
		return api.SMTPSettingsProjection{}, api.ErrSMTPSettingsInvalid
	}
	stores, err := application.base.bindOperation(ctx, scope, PermissionSMTPSettingsRead)
	if err != nil {
		return api.SMTPSettingsProjection{}, mapSMTPSettingsError(err)
	}
	config, err := stores.Providers().SelectProvider(ctx, domain.ChannelEmail, domain.ID{})
	if errors.Is(err, ports.ErrProviderNotFound) {
		return api.SMTPSettingsProjection{}, nil
	}
	if err != nil || config.Kind != ports.ProviderSMTP || !config.Enabled {
		return api.SMTPSettingsProjection{}, api.ErrSMTPSettingsUnavailable
	}
	return api.SMTPSettingsProjection{Configured: true,
		Host: config.Metadata.SMTPHost, Port: config.Metadata.SMTPPort,
		Username: config.Metadata.SMTPUsername, From: config.Metadata.FromAddress,
		PasswordConfigured: config.SecretID.Valid()}, nil
}

func mapSMTPSettingsError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid):
		return api.ErrSMTPSettingsInvalid
	case errors.Is(err, ErrForbidden):
		return api.ErrSMTPSettingsForbidden
	case errors.Is(err, ErrConflict):
		return api.ErrSMTPSettingsConflict
	case errors.Is(err, ErrNotFound):
		return api.ErrSMTPSettingsNotFound
	default:
		return api.ErrSMTPSettingsUnavailable
	}
}
