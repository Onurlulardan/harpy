package dashboardexport

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) DescribeContent(
	ctx context.Context, scope platformtx.TenantScope, dashboardID, exportID uuid.UUID,
) (StoredContent, error) {
	if !validUUIDv7(dashboardID) || !validUUIDv7(exportID) {
		return StoredContent{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationContent)
	if err != nil {
		return StoredContent{}, err
	}
	content, err := store.DescribeContent(ctx, dashboardID, exportID, application.clock.Now())
	if err != nil {
		return StoredContent{}, err
	}
	if content.ExportID != exportID || content.DashboardID != dashboardID ||
		content.Format.ContentType() == "" || content.ObjectVersion == "" ||
		content.Draft.PlaintextBytes < 1 || content.Draft.ObjectKey == "" ||
		content.Draft.Envelope.Validate() != nil {
		return StoredContent{}, ErrUnavailable
	}
	return content, nil
}

func (application *Application) OpenContent(
	ctx context.Context, stored StoredContent,
) (Content, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!stored.ExpiresAt.After(application.clock.Now()) {
		return Content{}, ErrExpired
	}
	reader, err := application.content.Open(ctx, stored)
	if err != nil || reader == nil {
		return Content{}, ErrUnavailable
	}
	return Content{Reader: reader, Length: stored.Draft.PlaintextBytes,
		ContentType: stored.Format.ContentType(), Filename: stored.Draft.Filename}, nil
}
