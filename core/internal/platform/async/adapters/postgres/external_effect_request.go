package postgres

import (
	"fmt"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/google/uuid"
)

type ExternalEffectRequestSpec struct {
	effectID      uuid.UUID
	provider      string
	operation     string
	correlationID string
	causationID   uuid.UUID
}

func NewExternalEffectRequestSpec(
	effectID uuid.UUID,
	provider string,
	operation string,
	correlationID string,
	causationID uuid.UUID,
) (ExternalEffectRequestSpec, error) {
	spec := ExternalEffectRequestSpec{
		effectID: effectID, provider: provider, operation: operation,
		correlationID: correlationID, causationID: causationID,
	}
	if effectID == uuid.Nil || causationID == uuid.Nil {
		return ExternalEffectRequestSpec{}, ErrExternalEffectInvalid
	}
	return spec, nil
}

func (spec ExternalEffectRequestSpec) Provider() string  { return spec.provider }
func (spec ExternalEffectRequestSpec) Operation() string { return spec.operation }

func buildExternalEffectRequest(
	claim JobClaim,
	spec ExternalEffectRequestSpec,
	requestedAt time.Time,
) (effectsv1.Request, error) {
	if claim.effectClass != "external" || !claim.hasEffectKey ||
		spec.effectID == uuid.Nil || spec.causationID == uuid.Nil || requestedAt.IsZero() {
		return effectsv1.Request{}, ErrExternalEffectInvalid
	}
	var tenantID *string
	if claim.tenantScoped {
		value := claim.tenantID.String()
		tenantID = &value
	}
	authority := effectsv1.Authority{
		ScopeKind: effectsv1.ScopeKind(claim.authority.scopeKind),
		ScopeID:   claim.authority.scopeID.String(), GroupID: claim.authority.groupID,
		GroupGeneration:   claim.authority.groupGeneration,
		Domain:            claim.authority.domain,
		CutoverGeneration: claim.authority.generation,
		WriterOrigin:      effectsv1.WriterV2,
	}
	request, err := effectsv1.NewRequest(
		spec.effectID.String(), claim.kind, claim.payloadRevision,
		tenantID, authority, spec.provider, spec.operation, claim.effectKey,
		spec.correlationID, spec.causationID.String(), claim.payloadReference,
		claim.payloadSHA256, requestedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return effectsv1.Request{}, fmt.Errorf("build external effect request: %w", err)
	}
	return request, nil
}
