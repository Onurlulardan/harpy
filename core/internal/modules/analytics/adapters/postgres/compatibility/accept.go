package compatibility

import (
	"context"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func (store *Store) Accept(
	ctx context.Context, authority compatibilityapp.AdoptionAuthority,
	receipt compatibilityapp.Acceptance,
) (compatibilityapp.Acceptance, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return compatibilityapp.Acceptance{}, false, compatibilityapp.ErrUnavailable
	}
	result := compatibilityapp.Acceptance{}
	var root []byte
	var replay bool
	err := store.database.QueryRow(ctx, acceptSQL, authority.SchemaName,
		authority.AuthorityID, authority.TenantID, receipt.RunID,
		int64(receipt.CutoverGeneration), int64(receipt.SourceCount),
		int64(receipt.AdoptedCount), int64(receipt.RetiredCount),
		receipt.RootSHA256[:], receipt.AcceptedAt).Scan(&result.RunID, &result.TenantID,
		&result.CutoverGeneration, &result.SourceCount, &result.AdoptedCount,
		&result.RetiredCount, &root, &result.AcceptedAt, &replay)
	if err != nil {
		return compatibilityapp.Acceptance{}, false, mapError(err)
	}
	if len(root) != 32 {
		return compatibilityapp.Acceptance{}, false, compatibilityapp.ErrUnavailable
	}
	copy(result.RootSHA256[:], root)
	return result, replay, nil
}
