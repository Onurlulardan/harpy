package resource

import (
	"context"
	"crypto/sha256"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row

func (verifier *Verifier) VerifyImmutable(
	ctx context.Context, scope platformtx.TenantScope,
	kind publicationapp.ResourceKind, versionID uuid.UUID,
) ([32]byte, error) {
	if verifier == nil || !kind.Valid() || !validUUIDv7(versionID) {
		return [32]byte{}, publicationapp.ErrInvalid
	}
	digest, err := verifier.read(ctx, scope, string(kind), versionID)
	if err != nil {
		return [32]byte{}, publicationapp.ErrConflict
	}
	return digest, nil
}

func (verifier *Verifier) VerifyPublished(
	ctx context.Context, scope platformtx.TenantScope, versions []uuid.UUID,
) ([32]byte, error) {
	if verifier == nil || len(versions) == 0 || len(versions) > 30 {
		return [32]byte{}, dashboardapp.ErrInvalid
	}
	return verifier.readMany(ctx, scope, versions)
}

func (verifier *Verifier) read(
	ctx context.Context, scope platformtx.TenantScope, kind string, versionID uuid.UUID,
) ([32]byte, error) {
	query, err := bindQuery(ctx, scope)
	if err != nil {
		return [32]byte{}, err
	}
	var stored []byte
	if err = query(ctx, manifestSQL, kind, versionID).Scan(&stored); err != nil || len(stored) != 32 {
		return [32]byte{}, errResourceUnavailable
	}
	var digest [32]byte
	copy(digest[:], stored)
	return digest, nil
}

func (verifier *Verifier) readMany(
	ctx context.Context, scope platformtx.TenantScope, versions []uuid.UUID,
) ([32]byte, error) {
	seen := make(map[uuid.UUID]struct{}, len(versions))
	digest := sha256.New()
	for _, versionID := range versions {
		if !validUUIDv7(versionID) {
			return [32]byte{}, dashboardapp.ErrInvalid
		}
		if _, duplicate := seen[versionID]; duplicate {
			return [32]byte{}, dashboardapp.ErrInvalid
		}
		seen[versionID] = struct{}{}
		manifest, err := verifier.read(ctx, scope, "query", versionID)
		if err != nil {
			return [32]byte{}, dashboardapp.ErrConflict
		}
		_, _ = digest.Write(versionID[:])
		_, _ = digest.Write(manifest[:])
	}
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result, nil
}

func bindQuery(
	ctx context.Context, scope platformtx.TenantScope,
) (queryRow, error) {
	if ctx == nil || ctx.Err() != nil || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, errResourceUnavailable
	}
	binder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) queryRow {
			return func(call context.Context, statement string, arguments ...any) row {
				return database.QueryRow(call, statement, arguments...)
			}
		})
	if err != nil {
		return nil, errResourceUnavailable
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, errResourceUnavailable
	}
	return bound, nil
}
