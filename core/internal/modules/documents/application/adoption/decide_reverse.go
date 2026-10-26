package adoption

import "context"

func (application *Application) DecideReverse(ctx context.Context,
	decision ReverseDecision,
) error {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!decision.ReconciliationID.UUIDv7() || !decision.ApprovedBy.Valid() ||
		decision.OccurredAt.IsZero() {
		return ErrInvalid
	}
	return mapError(application.store.DecideReverse(ctx, decision))
}
