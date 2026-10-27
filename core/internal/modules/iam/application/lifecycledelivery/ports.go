package lifecycledelivery

import (
	"context"

	"github.com/google/uuid"
)

type ClaimDecision struct {
	Artifact *Artifact
	Replay   *StoredResult
	Held     bool
}

func (decision ClaimDecision) Validate(tenantID uuid.UUID) error {
	count := 0
	if decision.Artifact != nil {
		count++
	}
	if decision.Replay != nil {
		count++
	}
	if decision.Held {
		count++
	}
	if count != 1 || tenantID == uuid.Nil {
		return ErrInvalid
	}
	if decision.Artifact != nil && decision.Artifact.Validate(tenantID) != nil {
		return ErrInvalid
	}
	if decision.Replay != nil && decision.Replay.Validate(tenantID) != nil {
		return ErrInvalid
	}
	return nil
}

type Store interface {
	Claim(context.Context, ClaimCommand) (ClaimDecision, error)
	Release(context.Context, EffectReference) error
	MarkUnknown(context.Context, EffectReference) error
	Finalize(context.Context, FinalizeCommand) (StoredResult, error)
}
