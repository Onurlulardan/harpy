package catalogquery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (queries *Queries) MaterializeLegacyIdentity(
	ctx context.Context,
	kind PlatformCatalogKind,
	legacyID uuid.UUID,
) (LegacyIdentityRecord, error) {
	if !queries.usable(ctx) || !validPlatformKind(kind) || legacyID == uuid.Nil {
		return LegacyIdentityRecord{}, ErrRequest
	}
	record, found, err := queries.store.MaterializeLegacyIdentity(
		ctx, LegacyIdentityStoreQuery{Kind: kind, LegacyID: legacyID},
	)
	if err != nil {
		return LegacyIdentityRecord{}, fmt.Errorf("materialize legacy billing identity: %w", err)
	}
	if !found {
		return LegacyIdentityRecord{}, ErrNotFound
	}
	if record.Kind != kind {
		return LegacyIdentityRecord{}, ErrRequest
	}
	if record.LegacyID != legacyID && record.CanonicalID != legacyID ||
		!validV7(record.CanonicalID) {
		return LegacyIdentityRecord{}, ErrResult
	}
	return record, nil
}
