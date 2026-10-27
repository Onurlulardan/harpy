package v1cutover

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func (store *Store) DisableV1(ctx context.Context, expected app.Authority,
	evidence app.Evidence,
) (app.Authority, error) {
	if !store.usable(ctx) || !expected.Valid() || evidence.ID == [16]byte{} ||
		evidence.Inventory.Validate() != nil || !evidence.Inventory.Zero() {
		return app.Authority{}, app.ErrInvalid
	}
	value, err := scanAuthority(store.queryRow(ctx, fenceSQL,
		expected.Generation, expected.RowVersion, evidence.ID,
		evidence.Inventory.RunID, evidence.Inventory.SourceEvidenceDigest[:]))
	if err != nil {
		return app.Authority{}, app.ErrNotReady
	}
	return value, nil
}
