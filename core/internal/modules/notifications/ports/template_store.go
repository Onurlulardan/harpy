package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type TemplateStore interface {
	LoadPublishedTemplate(context.Context, domain.ID, []string) (domain.PublishedTemplate, error)
	AddTemplate(context.Context, TemplateDraft) error
	StoreTemplateVersion(context.Context, TemplateVersion) error
	PublishTemplate(context.Context, domain.ID, domain.ID, int64) error
}

type TemplateDraft struct {
	ID               domain.ID
	NotificationType string
	Revision         int64
}

type TemplateVersion struct {
	Template domain.PublishedTemplate
	Digest   [32]byte
}
