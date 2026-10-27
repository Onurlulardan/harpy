package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) RecordHit(ctx context.Context,
	hit adoptionapp.CompatibilityHit,
) (adoptionapp.CompatibilityHit, error) {
	if !storage.usable(ctx) || hit.RouteContract == "" || hit.ConsumerID == "" ||
		hit.BuildID == "" || hit.LastHitAt.IsZero() {
		return adoptionapp.CompatibilityHit{}, adoptionapp.ErrInvalid
	}
	var result adoptionapp.CompatibilityHit
	var count int64
	err := storage.queryRow(ctx, recordHitSQL, hit.RouteContract, hit.ConsumerID,
		hit.BuildID, hit.LastHitAt).Scan(&result.RouteContract, &result.ConsumerID,
		&result.BuildID, &count, &result.FirstHitAt, &result.LastHitAt)
	if err != nil || count < 1 {
		return adoptionapp.CompatibilityHit{}, adoptionapp.ErrConflict
	}
	result.HitCount = uint64(count)
	return result, nil
}
