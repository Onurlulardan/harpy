package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type DatabaseRow interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) DatabaseRow
	Execute  func(context.Context, string, ...any) error
	Allowed  func(context.Context) bool
	NotFound func(error) bool
}

func NewJobStore(
	tenantID uuid.UUID,
	access DatabaseAccess,
	effects Effects,
) (subscriptionapp.Store, error) {
	return newWorkerStore(tenantID, access, effects, modeJob)
}

func newWorkerStore(
	tenantID uuid.UUID,
	access DatabaseAccess,
	effects Effects,
	mode bindingMode,
) (subscriptionapp.Store, error) {
	if tenantID == uuid.Nil || access.QueryRow == nil || access.Execute == nil ||
		access.Allowed == nil || access.NotFound == nil || effects == nil {
		return nil, subscriptionapp.ErrConfiguration
	}
	return &store{
		tenantID: tenantID, mode: mode, effects: effects,
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return access.QueryRow(ctx, statement, arguments...)
		},
		execute: access.Execute, allowed: access.Allowed,
		notFound: access.NotFound,
	}, nil
}
