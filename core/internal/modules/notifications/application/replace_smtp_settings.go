package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *SMTPSettingsApplication) ReplaceSMTPSettings(ctx context.Context,
	scope api.Scope, request api.SMTPSettingsReplaceRequest,
) (api.SMTPSettingsReplaceResult, error) {
	defer clear(request.Password)
	request, valid := normalizeSMTPReplace(request)
	if application == nil || ctx == nil || ctx.Err() != nil || !valid {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsInvalid
	}
	stores, err := application.base.bindOperation(ctx, scope, PermissionSMTPSettingsWrite)
	if err != nil {
		return api.SMTPSettingsReplaceResult{}, mapSMTPSettingsError(err)
	}
	settings := stores.SMTPSettings()
	if settings == nil || settings.LockSMTPSettings(ctx) != nil {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsUnavailable
	}
	digest := smtpReplaceDigest(request)
	if receipt, found, findErr := settings.FindSMTPSettingsReceipt(
		ctx, request.IdempotencyKey); findErr != nil {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsUnavailable
	} else if found {
		return replaySMTPSettings(receipt, request, digest)
	}
	latest, err := settings.LatestSMTPSettingsRevision(ctx)
	if err != nil {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsUnavailable
	}
	if latest > request.ExpectedSettingsRevision {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsConflict
	}
	config, err := application.applySMTPProvider(ctx, stores, request)
	if err != nil {
		return api.SMTPSettingsReplaceResult{}, mapSMTPSettingsError(err)
	}
	actor := scopeActor(scope)
	if !actor.Valid() || stores.Evidence().AppendAudit(ctx,
		"notifications.provider.smtp_settings.replace", config.ID,
		actor, config.Revision) != nil {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsUnavailable
	}
	receipt := ports.SMTPSettingsReceipt{IdempotencyKey: request.IdempotencyKey,
		RequestDigest: digest, ExpectedSettingsRevision: request.ExpectedSettingsRevision,
		ResultSettingsRevision: request.ResultSettingsRevision,
		ProviderConfigID:       config.ID, ProviderRevision: config.Revision}
	if err = settings.InsertSMTPSettingsReceipt(ctx, receipt); err != nil {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsUnavailable
	}
	return api.SMTPSettingsReplaceResult{
		ResultSettingsRevision: request.ResultSettingsRevision}, nil
}

func replaySMTPSettings(receipt ports.SMTPSettingsReceipt,
	request api.SMTPSettingsReplaceRequest, digest [32]byte,
) (api.SMTPSettingsReplaceResult, error) {
	if receipt.RequestDigest != digest ||
		receipt.ExpectedSettingsRevision != request.ExpectedSettingsRevision ||
		receipt.ResultSettingsRevision != request.ResultSettingsRevision ||
		!receipt.ProviderConfigID.Valid() || receipt.ProviderRevision < 1 {
		return api.SMTPSettingsReplaceResult{}, api.ErrSMTPSettingsConflict
	}
	return api.SMTPSettingsReplaceResult{
		ResultSettingsRevision: receipt.ResultSettingsRevision, Replay: true}, nil
}
