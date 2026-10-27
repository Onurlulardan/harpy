package api

import (
	"context"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PresentationFieldCatalog = presentation.FieldCatalog
type PresentationDocument = presentation.Document

type PublishPresentationInput struct {
	RevisionID          string
	ExpectedRowRevision int64
	Catalog             presentation.FieldCatalog
}

type ScopedResolver interface {
	ResolvePresentation(
		context.Context,
		platformtx.TenantScope,
		ResourceKind,
		presentation.FieldCatalog,
	) (presentation.Document, error)
}

type ScopedPublisher interface {
	PublishPresentation(
		context.Context,
		platformtx.TenantScope,
		PublishPresentationInput,
	) (Revision, error)
}
