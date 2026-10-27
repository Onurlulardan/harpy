package compatibility

import (
	"context"
)

func (application *Application) Accept(
	ctx context.Context, authority AdoptionAuthority, receipt Acceptance,
) (Acceptance, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validAuthority(authority) || !validUUID(receipt.RunID) ||
		receipt.TenantID != authority.TenantID ||
		receipt.CutoverGeneration == 0 || receipt.SourceCount !=
		receipt.AdoptedCount+receipt.RetiredCount || receipt.RootSHA256 == [32]byte{} {
		return Acceptance{}, ErrInvalid
	}
	receipt.AcceptedAt = application.clock.Now()
	if !validTime(receipt.AcceptedAt) {
		return Acceptance{}, ErrUnavailable
	}
	result, _, err := application.adoption.Accept(ctx, authority, receipt)
	if err != nil {
		return Acceptance{}, err
	}
	if result != receipt {
		return Acceptance{}, ErrUnavailable
	}
	return result, nil
}
