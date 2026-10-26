package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func (storage *store) MaterializeLegacyIdentity(
	ctx context.Context,
	query catalogapp.LegacyIdentityStoreQuery,
) (catalogapp.LegacyIdentityRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.LegacyIdentityRecord{}, false, catalogapp.ErrConfiguration
	}
	rows, err := storage.query(ctx, materializeLegacyIdentitySQL, query.Kind, query.LegacyID)
	if err != nil {
		return catalogapp.LegacyIdentityRecord{}, false,
			fmt.Errorf("query legacy billing identity: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return catalogapp.LegacyIdentityRecord{}, false, err
		}
		return catalogapp.LegacyIdentityRecord{}, false, nil
	}
	var result catalogapp.LegacyIdentityRecord
	if err = rows.Scan(&result.LegacyID, &result.CanonicalID, &result.Kind); err != nil {
		return catalogapp.LegacyIdentityRecord{}, false, err
	}
	if rows.Next() || rows.Err() != nil {
		return catalogapp.LegacyIdentityRecord{}, false,
			fmt.Errorf("query legacy billing identity: invalid cardinality")
	}
	return result, true, nil
}
