package customerlifecycle

import (
	"context"
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

func (store *Store) Mutate(ctx context.Context, request lifecycle.MutationStoreRequest) (
	lifecycle.MutationStoreResult, error,
) {
	if !store.usable(ctx) {
		return lifecycle.MutationStoreResult{}, lifecycle.ErrInvalid
	}
	var outcome string
	var replay bool
	row := store.queryRow(ctx, mutateSQL, request.EnvironmentID, request.ControlTenantID,
		request.ActorID, request.TargetTenantID, request.Kind, nullString(request.Name),
		request.ExpectedRevision, request.IdempotencyKey, request.RequestSHA256[:], request.Now)
	tenant, err := scanMutationTenant(row, store.classify, &outcome, &replay)
	if err != nil {
		return lifecycle.MutationStoreResult{}, err
	}
	return lifecycle.MutationStoreResult{Tenant: tenant, Outcome: outcome, Replay: replay}, nil
}

func scanMutationTenant(row Row, classify func(error) error, prefix ...any) (
	lifecycle.Tenant, error,
) {
	var id *uuid.UUID
	var name, slug, status *string
	var revision, settingsRevision *int64
	var createdAt, updatedAt, offboardedAt, purgedAt *time.Time
	targets := append(prefix, &id, &name, &slug, &status, &revision, &settingsRevision,
		&createdAt, &updatedAt, &offboardedAt, &purgedAt)
	if err := row.Scan(targets...); err != nil {
		return lifecycle.Tenant{}, classify(err)
	}
	if id == nil {
		return lifecycle.Tenant{}, nil
	}
	if name == nil || slug == nil || status == nil || revision == nil ||
		settingsRevision == nil || createdAt == nil || updatedAt == nil {
		return lifecycle.Tenant{}, lifecycle.ErrUnavailable
	}
	result := lifecycle.Tenant{ID: *id, Name: *name, Slug: *slug, Status: *status,
		Revision: *revision, SettingsRevision: *settingsRevision,
		CreatedAt: createdAt.UTC(), UpdatedAt: updatedAt.UTC()}
	if offboardedAt != nil {
		result.OffboardedAt = offboardedAt.UTC()
	}
	if purgedAt != nil {
		result.PurgedAt = purgedAt.UTC()
	}
	return result, nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
