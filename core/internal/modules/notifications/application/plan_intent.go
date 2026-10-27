package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *Application) planIntent(ctx context.Context, stores ports.Stores,
	recipient ports.Recipient, request api.IntentRequest,
	template domain.PublishedTemplate, candidate domain.EffectCandidate,
	providers resolvedProviders,
) (intentPlan, error) {
	now := application.clock.Now().UTC()
	messageID, err := application.ids.New(now)
	if err != nil || !messageID.IsV7() {
		return intentPlan{}, ErrUnavailable
	}
	revision, err := stores.Messages().NextInboxRevision(ctx, recipient.PrincipalID)
	if err != nil || revision < 1 {
		return intentPlan{}, ErrUnavailable
	}
	data, err := renderData(request.Data)
	if err != nil {
		return intentPlan{}, err
	}
	rendered, err := template.Render(data)
	if err != nil {
		return intentPlan{}, ErrInvalid
	}
	message, err := domain.NewMessage(domain.Message{ID: messageID,
		RecipientID: recipient.PrincipalID, TemplateVersionID: template.VersionID,
		Producer: request.Producer, NotificationType: request.NotificationType,
		Category: category(request.Category), TemplateKind: template.Kind,
		Locale:  rendered.Locale,
		Subject: rendered.Subject, Body: rendered.Body, DataDigest: request.DataDigest,
		SubjectDigest: hashContent(rendered.Subject), BodyDigest: hashContent(rendered.Body),
		Revision: 1, InboxRevision: revision, CreatedAt: now})
	if err != nil {
		return intentPlan{}, ErrInvalid
	}
	receipt, err := domain.NewEffectReceipt(candidate, messageID, now, now.AddDate(1, 0, 0))
	if err != nil {
		return intentPlan{}, ErrInvalid
	}
	plan := intentPlan{message: message, receipt: receipt}
	for _, selected := range request.Channels {
		if selected == api.ChannelInApp {
			plan.hint = &ports.RealtimeHint{MessageID: messageID,
				RecipientID: recipient.PrincipalID, InboxRevision: revision}
			continue
		}
		planned, planErr := application.planDelivery(ctx, stores, recipient,
			request, messageID, selected, now, template.MandatoryEmail,
			providers[channel(selected)])
		if planErr != nil {
			return intentPlan{}, planErr
		}
		plan.deliveries = append(plan.deliveries, planned)
	}
	return plan, nil
}

func (application *Application) nextID(at time.Time) (domain.ID, error) {
	value, err := application.ids.New(at)
	if err != nil || !value.IsV7() {
		return domain.ID{}, ErrUnavailable
	}
	return value, nil
}
