package systemtransfer

import "context"

type AttemptState struct {
	Attempt      Attempt
	Source       ExactImportSource
	NextRowIndex int64
	Processed    int64
	Finished     bool
}

func (state AttemptState) Validate(claim AttemptClaim) error {
	if claim.Validate() != nil || state.Attempt.Validate() != nil ||
		state.Source.Validate() != nil || state.NextRowIndex != state.Processed+1 ||
		state.Processed < 0 || state.Processed > 1_000_000 ||
		state.Attempt.ID != claim.ImportID || state.Attempt.JobID != claim.JobID ||
		state.Attempt.AttemptID != claim.AttemptID ||
		state.Attempt.LeaseOwnerID != claim.LeaseOwnerID ||
		state.Attempt.LeaseGeneration != claim.LeaseGeneration ||
		state.Attempt.TenantID != claim.TenantID ||
		state.Attempt.Resource != claim.Resource ||
		state.Source.TenantID != claim.TenantID ||
		state.Source.ActorID != state.Attempt.ActorID ||
		state.Source.Resource != claim.Resource {
		return ErrWorkerState
	}
	return nil
}

type AttemptLoader interface {
	LoadAttempt(context.Context, AttemptClaim) (AttemptState, error)
}
