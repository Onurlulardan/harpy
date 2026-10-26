package recoveryanchorpostgres

import (
	"context"
	"fmt"
)

func (store *authorityFinalizerStore) FinalizeAuthority(
	ctx context.Context,
	operationID string,
) error {
	if store == nil || store.transactions == nil ||
		ctx == nil || operationID == "" {
		return fmt.Errorf("IAM recovery anchor authority finalizer is unavailable")
	}
	return store.transactions.Within(
		ctx,
		func(database authorityFinalizerDatabase) error {
			var completed bool
			if err := database.QueryRow(
				ctx, finalizeAuthoritySQL, operationID,
			).Scan(&completed); err != nil {
				return fmt.Errorf("finalize IAM recovery anchor authority: %w", err)
			}
			if !completed {
				return fmt.Errorf("IAM recovery anchor authority did not complete")
			}
			return nil
		},
	)
}
