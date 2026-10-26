package adoption

import "context"

func (application *Application) RequestReverse(ctx context.Context,
	request ReverseRequest,
) error {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!request.ReconciliationID.UUIDv7() || !request.RequestedBy.Valid() ||
		request.SourceGeneration == 0 || request.TargetGeneration == 0 ||
		request.SourceSHA256 == [32]byte{} || request.TargetSHA256 == [32]byte{} ||
		request.OccurredAt.IsZero() {
		return ErrInvalid
	}
	return mapError(application.store.RequestReverse(ctx, request))
}
