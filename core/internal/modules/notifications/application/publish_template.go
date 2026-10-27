package application

import (
	"context"
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type PublishTemplateInput struct {
	TemplateID, VersionID domain.ID
	NotificationType      string
	Kind                  domain.TemplateKind
	Locale                string
	Subject, Body         string
	Variables             []string
	MandatoryEmail        bool
	ExpectedRevision      int64
}

func (application *Application) PublishTemplate(ctx context.Context, scope api.Scope,
	input PublishTemplateInput,
) (domain.PublishedTemplate, error) {
	if input.ExpectedRevision < 1 {
		return domain.PublishedTemplate{}, ErrInvalid
	}
	stores, err := application.bindOperation(ctx, scope, PermissionSend)
	if err != nil {
		return domain.PublishedTemplate{}, err
	}
	if !input.VersionID.Valid() {
		input.VersionID, err = application.nextID(application.clock.Now())
		if err != nil {
			return domain.PublishedTemplate{}, err
		}
	}
	template := domain.PublishedTemplate{ID: input.TemplateID, VersionID: input.VersionID,
		NotificationType: input.NotificationType, Kind: input.Kind, Locale: input.Locale,
		SubjectPattern: input.Subject, BodyPattern: input.Body,
		Variables:      append([]string(nil), input.Variables...),
		MandatoryEmail: input.MandatoryEmail, Revision: input.ExpectedRevision}
	if !template.Valid() {
		return domain.PublishedTemplate{}, ErrInvalid
	}
	digest := sha256.Sum256([]byte(input.NotificationType + "\x00" + input.Locale +
		"\x00" + input.Subject + "\x00" + input.Body))
	if err = stores.Templates().StoreTemplateVersion(ctx,
		ports.TemplateVersion{Template: template, Digest: digest}); err != nil {
		return domain.PublishedTemplate{}, ErrConflict
	}
	if err = stores.Templates().PublishTemplate(ctx, input.TemplateID,
		input.VersionID, input.ExpectedRevision); err != nil {
		return domain.PublishedTemplate{}, ErrConflict
	}
	if err = stores.Evidence().AppendAudit(ctx, "notifications.template.publish",
		input.TemplateID, scopeActor(scope), input.ExpectedRevision+1); err != nil {
		return domain.PublishedTemplate{}, ErrUnavailable
	}
	return template, nil
}
