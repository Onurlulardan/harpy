package application

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestPublishTemplateRejectsExecutableMarkupAndPinsImmutableVersion(t *testing.T) {
	application, stores, scope, _ := newIntentFixture(t)
	draft, _ := application.AddTemplate(t.Context(), scope,
		CreateTemplateInput{NotificationType: "document.ready"})
	input := PublishTemplateInput{TemplateID: draft.ID, NotificationType: "document.ready",
		Kind: domain.TemplatePlainText, Locale: "tr", Subject: "Hazır {{name}}",
		Body: "Belge {{name}}", Variables: []string{"name"}, ExpectedRevision: 1}
	published, err := application.PublishTemplate(t.Context(), scope, input)
	if err != nil || !published.VersionID.IsV7() ||
		stores.createdTemplateVersion.Template.VersionID != published.VersionID {
		t.Fatalf("PublishTemplate() = %+v, %v", published, err)
	}
	input.Body = "{{if .admin}}secret{{end}}"
	if _, err = application.PublishTemplate(t.Context(), scope, input); err != ErrInvalid {
		t.Fatalf("executable template error = %v", err)
	}
}
