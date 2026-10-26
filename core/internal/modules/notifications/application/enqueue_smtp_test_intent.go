package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *SMTPSettingsApplication) EnqueueSMTPTest(ctx context.Context,
	scope api.Scope, request api.SMTPTestIntentRequest,
) (api.SMTPTestIntentResult, error) {
	request, valid := normalizeSMTPTest(request)
	if application == nil || ctx == nil || ctx.Err() != nil || !valid {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsInvalid
	}
	stores, err := application.base.bindOperation(ctx, scope, PermissionSMTPSettingsTest)
	if err != nil {
		return api.SMTPTestIntentResult{}, mapSMTPSettingsError(err)
	}
	settings := stores.SMTPSettings()
	if settings == nil || settings.LockSMTPSettings(ctx) != nil {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	}
	digest := smtpTestDigest(request)
	if existing, found, findErr := settings.FindSMTPTestIntent(
		ctx, request.IdempotencyKey); findErr != nil {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	} else if found {
		return replaySMTPTest(existing, request, digest)
	}
	latest, err := settings.LatestSMTPSettingsRevision(ctx)
	if err != nil {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	}
	if latest > request.ExpectedSettingsRevision {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsConflict
	}
	provider, err := stores.Providers().SelectProvider(ctx,
		domain.ChannelEmail, domain.ID{})
	if errors.Is(err, ports.ErrProviderNotFound) {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsNotFound
	}
	if err != nil || !provider.Enabled || provider.Kind != ports.ProviderSMTP ||
		provider.Metadata.SMTPUsername == "" || !provider.SecretID.Valid() {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	}
	operationID, err := application.base.nextID(application.base.clock.Now().UTC())
	if err != nil {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	}
	target := []byte(request.To)
	sealed, err := application.sealer.Seal(ctx, operationID, target,
		application.kekRevision)
	clear(target)
	if err != nil {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	}
	actor := scopeActor(scope)
	intent := ports.SMTPTestIntent{OperationID: operationID,
		ProviderVersionID: provider.VersionID, ActorID: actor,
		IdempotencyKey: request.IdempotencyKey, RequestDigest: digest,
		ExpectedSettingsRevision: request.ExpectedSettingsRevision,
		EncryptedTarget:          sealed, RequestedAt: request.RequestedAt.UTC()}
	if !actor.Valid() || settings.InsertSMTPTestIntent(ctx, intent) != nil ||
		stores.Evidence().AppendAudit(ctx, "notifications.provider.smtp_test.enqueue",
			operationID, actor, 1) != nil {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsUnavailable
	}
	return api.SMTPTestIntentResult{OperationID: wireID(operationID)}, nil
}

func replaySMTPTest(existing ports.SMTPTestIntent,
	request api.SMTPTestIntentRequest, digest [32]byte,
) (api.SMTPTestIntentResult, error) {
	if existing.RequestDigest != digest ||
		existing.ExpectedSettingsRevision != request.ExpectedSettingsRevision ||
		!existing.OperationID.IsV7() || !existing.ProviderVersionID.Valid() {
		return api.SMTPTestIntentResult{}, api.ErrSMTPSettingsConflict
	}
	return api.SMTPTestIntentResult{
		OperationID: wireID(existing.OperationID), Replay: true}, nil
}
