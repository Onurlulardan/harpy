package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) DecideReverse(ctx context.Context,
	decision adoptionapp.ReverseDecision,
) error {
	if !storage.usable(ctx) || !decision.ReconciliationID.UUIDv7() ||
		!decision.ApprovedBy.Valid() || decision.OccurredAt.IsZero() {
		return adoptionapp.ErrInvalid
	}
	var result bool
	err := storage.queryRow(ctx, decideReverseSQL, decision.ReconciliationID.String(),
		decision.ApprovedBy.String(), decision.Accept, decision.OccurredAt).Scan(&result)
	if err != nil || !result {
		return adoptionapp.ErrConflict
	}
	return nil
}
