package adoption

import (
	"context"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) RequestReverse(ctx context.Context,
	request adoptionapp.ReverseRequest,
) error {
	if !storage.usable(ctx) || !request.ReconciliationID.UUIDv7() ||
		request.SourceGeneration == 0 || request.TargetGeneration == 0 ||
		request.SourceSHA256 == [32]byte{} || request.TargetSHA256 == [32]byte{} ||
		!request.RequestedBy.Valid() || request.OccurredAt.IsZero() {
		return adoptionapp.ErrInvalid
	}
	var result bool
	err := storage.queryRow(ctx, requestReverseSQL, request.ReconciliationID.String(),
		int64(request.SourceGeneration), int64(request.TargetGeneration),
		int64(request.SourceCount), int64(request.TargetCount), request.SourceSHA256[:],
		request.TargetSHA256[:], request.RequestedBy.String(), request.OccurredAt).Scan(&result)
	if err != nil || !result {
		return adoptionapp.ErrConflict
	}
	return nil
}
