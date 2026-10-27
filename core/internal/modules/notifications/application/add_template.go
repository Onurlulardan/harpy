package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type CreateTemplateInput struct{ NotificationType string }

func (application *Application) AddTemplate(ctx context.Context, scope api.Scope,
	input CreateTemplateInput,
) (ports.TemplateDraft, error) {
	input.NotificationType = strings.TrimSpace(input.NotificationType)
	if input.NotificationType == "" || len(input.NotificationType) > 96 {
		return ports.TemplateDraft{}, ErrInvalid
	}
	stores, err := application.bindOperation(ctx, scope, PermissionSend)
	if err != nil {
		return ports.TemplateDraft{}, err
	}
	id, err := application.nextID(application.clock.Now())
	if err != nil {
		return ports.TemplateDraft{}, err
	}
	draft := ports.TemplateDraft{ID: id, NotificationType: input.NotificationType, Revision: 1}
	if err = stores.Templates().AddTemplate(ctx, draft); err != nil {
		return ports.TemplateDraft{}, ErrUnavailable
	}
	if err = stores.Evidence().AppendAudit(ctx, "notifications.template.create",
		id, scopeActor(scope), draft.Revision); err != nil {
		return ports.TemplateDraft{}, ErrUnavailable
	}
	return draft, nil
}
