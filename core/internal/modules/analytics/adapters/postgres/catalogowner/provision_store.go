package catalogowner

import (
	"context"
	"fmt"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	catalogprojection "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogprojection/api"
	"github.com/google/uuid"
)

type provisionStore struct {
	generation int64
	actorID    uuid.UUID
	queryRow   catalogprojection.QueryRow
	allowed    catalogprojection.ScopeGuard
}

func (store *provisionStore) Project(
	ctx context.Context, input catalogprojection.Projection,
) (catalogprojection.Snapshot, error) {
	materialization := input.Materialization
	if !store.usable(ctx) || input.ProposedReceiptID.Version() != 7 ||
		input.PackageID == uuid.Nil || input.ReleaseID.Version() != 7 ||
		input.Component.Owner != owner ||
		input.Component.Contract != catalogpackagev1.ContractAnalytics ||
		input.Component.Generation != store.generation ||
		(materialization.Kind != "ENDPOINT" && materialization.Kind != "DASHBOARD") ||
		input.RequestSHA256 == [32]byte{} || input.OccurredAt.IsZero() {
		return catalogprojection.Snapshot{}, catalogprojection.ErrInvalid
	}
	runtime, runtimeSHA256, err := buildRuntimePlan(
		materialization.Document, input.Component,
	)
	if err != nil {
		return catalogprojection.Snapshot{}, catalogprojection.ErrInvalid
	}
	snapshot, err := scanProvisionSnapshot(store.queryRow(
		ctx, provisionSQL,
		input.ProposedReceiptID, input.PackageID, input.PackageSlug,
		input.PackageVersion, input.ReleaseID, input.Component.ID,
		input.Component.Generation, input.Component.Contract,
		input.Component.ContentSHA256, materialization.Kind,
		materialization.ModuleSlug, materialization.Identity,
		materialization.SemanticSHA256[:], materialization.ItemCount,
		materialization.AuxiliaryCount, materialization.Document,
		runtime, runtimeSHA256[:], store.actorID, input.IdempotencyKey,
		input.RequestSHA256[:], input.OccurredAt,
	))
	if err != nil {
		return catalogprojection.Snapshot{}, fmt.Errorf(
			"execute Analytics catalog materialization: %w", err,
		)
	}
	return snapshot, nil
}

func (store *provisionStore) usable(ctx context.Context) bool {
	return store != nil && store.generation > 0 && store.actorID != uuid.Nil &&
		store.queryRow != nil && store.allowed != nil &&
		ctx != nil && ctx.Err() == nil && store.allowed(ctx)
}

var _ catalogprojection.Store = (*provisionStore)(nil)
