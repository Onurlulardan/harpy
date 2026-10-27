package tenantsettingspostgres

import (
	"context"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/google/uuid"
)

var uuidNil = uuid.Nil

func (storage *store) Update(ctx context.Context, scope settings.Scope,
	mutation settings.RegistryMutation,
) (settings.UpdateResult, error) {
	if !contextUsable(ctx) || scope.TenantID() == uuidNil || !validMutation(mutation) {
		return settings.UpdateResult{}, settings.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return settings.UpdateResult{}, err
	}
	var result settings.UpdateResult
	var outcome string
	err = bound.database.QueryRow(ctx, updateSQL, scope.TenantID(),
		nullableName(mutation.Name), mutation.ExpectedRevision,
		mutation.IdempotencyKey, mutation.RequestSHA256[:]).Scan(
		&result.Revision, &result.Replay, &outcome)
	if err != nil {
		return settings.UpdateResult{}, settings.ErrUnavailable
	}
	switch outcome {
	case "applied", "replay":
		if result.Revision < 1 || result.Replay != (outcome == "replay") {
			return settings.UpdateResult{}, settings.ErrUnavailable
		}
		return result, nil
	case "not_found":
		return settings.UpdateResult{}, settings.ErrNotFound
	case "revision_conflict", "idempotency_conflict":
		return settings.UpdateResult{}, settings.ErrConflict
	default:
		return settings.UpdateResult{}, settings.ErrUnavailable
	}
}
