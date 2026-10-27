package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *Application) RecordIntent(ctx context.Context, scope api.Scope,
	request api.IntentRequest,
) (api.IntentResult, error) {
	if application == nil || ctx == nil || !validScope(scope) || request.Validate() != nil {
		return api.IntentResult{}, api.ErrIntentInvalid
	}
	if err := application.authorizer.AuthorizeIntent(
		ctx, scope, request.Producer, request.NotificationType,
	); err != nil {
		return api.IntentResult{}, api.ErrForbidden
	}
	recipientID, err := domainID(request.RecipientPrincipalID)
	if err != nil {
		return api.IntentResult{}, api.ErrIntentInvalid
	}
	recipient, err := application.recipients.ResolveRecipient(ctx, scope, recipientID)
	if err != nil || recipient.PrincipalID != recipientID {
		return api.IntentResult{}, api.ErrRecipientNotFound
	}
	tenantID, _, _ := scopeIDs(scope)
	if recipient.TenantID != tenantID || recipient.AuthorizationVersion < 1 {
		return api.IntentResult{}, api.ErrRecipientNotFound
	}
	stores, err := application.binder.Bind(ctx, scope)
	if err != nil || stores == nil {
		return api.IntentResult{}, api.ErrUnavailable
	}
	result, err := application.recordIntent(ctx, stores, recipient, request, scopeActor(scope))
	return result, mapIntentError(err)
}

func (application *Application) recordIntent(ctx context.Context, stores ports.Stores,
	recipient ports.Recipient, request api.IntentRequest, actor domain.ID,
) (api.IntentResult, error) {
	templateID, err := domainID(request.TemplateVersionID)
	if err != nil {
		return api.IntentResult{}, ErrInvalid
	}
	template, err := stores.Templates().LoadPublishedTemplate(
		ctx, templateID, append([]string{request.LocaleInput}, recipient.SupportedLocales...),
	)
	if err != nil || template.VersionID != templateID ||
		template.NotificationType != request.NotificationType {
		return api.IntentResult{}, api.ErrTemplateNotFound
	}
	request.Channels = effectiveChannels(request, template)
	request.LocaleInput = template.Locale
	providers, err := selectProviders(ctx, stores, request)
	if err != nil {
		return api.IntentResult{}, err
	}
	candidate := domain.EffectCandidate{Producer: request.Producer,
		EffectKey: request.EffectKey, RecipientID: recipient.PrincipalID,
		CanonicalDigest: intentDigest(request, recipient.PrincipalID, providers)}
	if err = stores.Effects().LockEffect(ctx, candidate.Producer,
		candidate.KeyHash(), candidate.RecipientID); err != nil {
		return api.IntentResult{}, ErrUnavailable
	}
	if existing, found, findErr := stores.Effects().FindEffect(ctx,
		candidate.Producer, candidate.KeyHash(), candidate.RecipientID); findErr != nil {
		return api.IntentResult{}, ErrUnavailable
	} else if found {
		messageID, outcome, replayErr := existing.Replay(candidate)
		if replayErr != nil {
			return api.IntentResult{}, replayErr
		}
		message, getErr := stores.Messages().GetMessage(ctx, messageID, recipient.PrincipalID)
		if getErr != nil {
			return api.IntentResult{}, ErrUnavailable
		}
		return api.IntentResult{MessageID: wireID(messageID),
			InboxRevision: message.InboxRevision, Replay: true, Outcome: string(outcome)}, nil
	}
	plan, err := application.planIntent(ctx, stores, recipient, request,
		template, candidate, providers)
	if err != nil {
		return api.IntentResult{}, err
	}
	if err = persistIntent(ctx, stores, plan, actor); err != nil {
		return api.IntentResult{}, err
	}
	return api.IntentResult{MessageID: wireID(plan.message.ID),
		InboxRevision: plan.message.InboxRevision, Outcome: string(plan.receipt.Outcome)}, nil
}

func mapIntentError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrEffectReused):
		return api.ErrEffectKeyReused
	case errors.Is(err, ErrForbidden):
		return api.ErrForbidden
	case errors.Is(err, ErrInvalid):
		return api.ErrIntentInvalid
	default:
		return api.ErrUnavailable
	}
}

func scopeActor(scope api.Scope) domain.ID {
	_, value, _ := scopeIDs(scope)
	return value
}
