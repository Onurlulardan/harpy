package invitationdelivery

import (
	"time"

	"github.com/google/uuid"
)

type EffectReference struct {
	TenantID    uuid.UUID
	ArtifactRef string
	EffectID    uuid.UUID
	EffectKey   string
	OccurredAt  time.Time
}

func (reference EffectReference) Validate() error {
	return (ClaimCommand{
		TenantID: reference.TenantID, ArtifactRef: reference.ArtifactRef,
		EffectID: reference.EffectID, EffectKey: reference.EffectKey,
		MessageID: "validation@harpy.invalid", ClaimedAt: reference.OccurredAt,
	}).Validate()
}

type FinalizeCommand struct {
	EffectReference
	Result StoredResult
}

func (command FinalizeCommand) Validate() error {
	if command.EffectReference.Validate() != nil ||
		command.Result.Validate(command.TenantID) != nil ||
		command.Result.ArtifactRef != command.ArtifactRef ||
		command.Result.EffectKey != command.EffectKey {
		return ErrInvalid
	}
	return nil
}
