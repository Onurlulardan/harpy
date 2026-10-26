package invitationdelivery

import (
	"regexp"
	"time"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
)

var machineCode = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)

type ClaimCommand struct {
	TenantID    uuid.UUID
	ArtifactRef string
	EffectID    uuid.UUID
	EffectKey   string
	MessageID   string
	ClaimedAt   time.Time
}

func (command ClaimCommand) Validate() error {
	reference, err := deliveryartifactv1.ParseReference(command.ArtifactRef)
	if err != nil || reference.TenantID() != command.TenantID ||
		command.EffectID == uuid.Nil || len(command.EffectKey) < 1 ||
		len(command.EffectKey) > 512 || len(command.MessageID) < 3 ||
		len(command.MessageID) > 512 || command.ClaimedAt.IsZero() {
		return ErrInvalid
	}
	return nil
}

type Artifact struct {
	Reference string
	Recipient string
	Envelope  deliveryartifactv1.Envelope
	ExpiresAt time.Time
}

func (artifact Artifact) Validate(tenantID uuid.UUID) error {
	reference, err := deliveryartifactv1.ParseReference(artifact.Reference)
	if err != nil || reference.TenantID() != tenantID || artifact.Recipient == "" ||
		artifact.Envelope.Validate(tenantID) != nil || artifact.ExpiresAt.IsZero() ||
		reference.ArtifactID() != artifact.Envelope.ArtifactID {
		return ErrInvalid
	}
	return nil
}

type Outcome string

const (
	OutcomeDelivered Outcome = "delivered"
	OutcomePermanent Outcome = "permanent_failure"
)

func (outcome Outcome) Valid() bool {
	return outcome == OutcomeDelivered || outcome == OutcomePermanent
}

type StoredResult struct {
	ReceiptID         uuid.UUID
	ArtifactRef       string
	EffectKey         string
	MessageID         string
	Outcome           Outcome
	Code              string
	ProviderReference string
	ResultSHA256      [32]byte
	CompletedAt       time.Time
}

func (result StoredResult) Validate(tenantID uuid.UUID) error {
	reference, err := deliveryartifactv1.ParseReference(result.ArtifactRef)
	if err != nil || reference.TenantID() != tenantID || result.ReceiptID == uuid.Nil ||
		len(result.EffectKey) < 1 || len(result.EffectKey) > 512 ||
		len(result.MessageID) < 3 || len(result.MessageID) > 512 ||
		!result.Outcome.Valid() || !machineCode.MatchString(result.Code) ||
		len(result.ProviderReference) < 1 || len(result.ProviderReference) > 512 ||
		result.ResultSHA256 == [32]byte{} || result.CompletedAt.IsZero() {
		return ErrInvalid
	}
	return nil
}
