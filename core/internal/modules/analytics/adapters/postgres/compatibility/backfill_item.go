package compatibility

import (
	"context"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	"github.com/google/uuid"
)

func (store *Store) BackfillItem(
	ctx context.Context, authority compatibilityapp.AdoptionAuthority,
	item compatibilityapp.AdoptionItem,
) (compatibilityapp.AdoptionItem, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return compatibilityapp.AdoptionItem{}, false, compatibilityapp.ErrUnavailable
	}
	row := store.database.QueryRow(ctx, recordItemSQL, authority.SchemaName,
		authority.AuthorityID, authority.TenantID, item.RunID, item.LegacyEndpointID,
		item.SourceSHA256[:], string(item.Disposition), nullableText(item.ReasonCode),
		nullableUUID(item.QueryID), nullableUUID(item.QueryVersionID),
		nullableUUID(item.PublicationID), nullableDigest(item.TargetSHA256), item.CreatedAt)
	result := compatibilityapp.AdoptionItem{}
	var source, target []byte
	var disposition string
	var reason *string
	var queryID, queryVersionID, publicationID *uuid.UUID
	var replay bool
	err := row.Scan(&result.RunID, &result.LegacyEndpointID, &source, &disposition,
		&reason, &queryID, &queryVersionID, &publicationID, &target, &result.CreatedAt, &replay)
	if err != nil {
		return compatibilityapp.AdoptionItem{}, false, mapError(err)
	}
	if len(source) != 32 || (len(target) != 0 && len(target) != 32) {
		return compatibilityapp.AdoptionItem{}, false, compatibilityapp.ErrUnavailable
	}
	copy(result.SourceSHA256[:], source)
	copy(result.TargetSHA256[:], target)
	result.Disposition = compatibilityapp.Disposition(disposition)
	if reason != nil {
		result.ReasonCode = *reason
	}
	assignUUID(&result.QueryID, queryID)
	assignUUID(&result.QueryVersionID, queryVersionID)
	assignUUID(&result.PublicationID, publicationID)
	return result, replay, nil
}
