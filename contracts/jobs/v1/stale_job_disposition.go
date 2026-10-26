package jobsv1

import (
	"encoding/json"
	"fmt"
	"time"
)

const StaleJobDispositionContractVersion = "stale-job-disposition.v1"

type StaleJobDecision string

const (
	StaleJobAdopt  StaleJobDecision = "adopt"
	StaleJobCancel StaleJobDecision = "cancel"
)

type StaleJobPayloadIdentity struct {
	Revision  int     `json:"revision"`
	SHA256    string  `json:"sha256"`
	Reference string  `json:"reference"`
	DedupeKey string  `json:"dedupe_key"`
	EffectKey *string `json:"effect_idempotency_key"`
}

type StaleJobDisposition struct {
	ContractVersion  string                   `json:"contract_version"`
	EvidenceRootID   string                   `json:"evidence_root_id"`
	DispositionID    string                   `json:"disposition_id"`
	SourceJobID      string                   `json:"source_job_id"`
	Decision         StaleJobDecision         `json:"decision"`
	ResultJobID      *string                  `json:"result_job_id"`
	Kind             string                   `json:"kind"`
	EffectClass      EffectClass              `json:"effect_class"`
	MaximumAttempts  int                      `json:"maximum_attempts"`
	Priority         int32                    `json:"priority"`
	OldAuthority     Authority                `json:"old_authority"`
	CurrentAuthority Authority                `json:"current_authority"`
	SourcePayload    StaleJobPayloadIdentity  `json:"source_payload"`
	ResultPayload    *StaleJobPayloadIdentity `json:"result_payload"`
	SourceRunAt      string                   `json:"source_run_at"`
	ResultRunAt      *string                  `json:"result_run_at"`
	Reason           string                   `json:"reason"`
	TicketReference  string                   `json:"ticket_reference"`
	ActorID          string                   `json:"actor_id"`
	CausationID      string                   `json:"causation_id"`
	OccurredAt       string                   `json:"occurred_at"`
}

func NewStaleJobDisposition(value StaleJobDisposition) (StaleJobDisposition, error) {
	value.ContractVersion = StaleJobDispositionContractVersion
	value.ResultJobID = cloneOptional(value.ResultJobID)
	value.ResultRunAt = cloneOptional(value.ResultRunAt)
	value.SourcePayload = cloneStalePayload(value.SourcePayload)
	if value.ResultPayload != nil {
		result := cloneStalePayload(*value.ResultPayload)
		value.ResultPayload = &result
	}
	if err := value.Validate(); err != nil {
		return StaleJobDisposition{}, err
	}
	return value, nil
}

func (value StaleJobDisposition) Validate() error {
	if value.ContractVersion != StaleJobDispositionContractVersion ||
		!validUUID(value.EvidenceRootID) || !validUUID(value.DispositionID) ||
		!validUUID(value.SourceJobID) || value.EvidenceRootID == value.DispositionID ||
		value.EvidenceRootID == value.SourceJobID || value.DispositionID == value.SourceJobID {
		return fmt.Errorf("stale job disposition identity is invalid")
	}
	if !validKind(value.Kind) || !value.EffectClass.valid() ||
		value.MaximumAttempts < MinimumMaximumAttempts ||
		value.MaximumAttempts > MaximumMaximumAttempts {
		return fmt.Errorf("stale job policy is invalid")
	}
	if err := value.OldAuthority.Validate(); err != nil {
		return err
	}
	if err := value.CurrentAuthority.Validate(); err != nil {
		return err
	}
	if !sameStaleAuthorityIdentity(value.OldAuthority, value.CurrentAuthority) ||
		value.CurrentAuthority.WriterOrigin != WriterV2 ||
		value.CurrentAuthority.GroupGeneration < value.OldAuthority.GroupGeneration ||
		value.CurrentAuthority.CutoverGeneration < value.OldAuthority.CutoverGeneration ||
		(value.CurrentAuthority.GroupGeneration == value.OldAuthority.GroupGeneration &&
			value.CurrentAuthority.CutoverGeneration == value.OldAuthority.CutoverGeneration) {
		return fmt.Errorf("stale job authority transition is invalid")
	}
	if err := value.SourcePayload.validate(); err != nil {
		return err
	}
	if value.EffectClass == EffectDatabaseTransactional && value.SourcePayload.EffectKey != nil ||
		value.EffectClass == EffectExternal && value.SourcePayload.EffectKey == nil {
		return fmt.Errorf("stale job effect policy is invalid")
	}
	if !validText(value.Reason, MaximumReasonBytes) ||
		!validText(value.TicketReference, MaximumTicketBytes) ||
		!validUUID(value.ActorID) || !validUUID(value.CausationID) {
		return fmt.Errorf("stale job disposition audit is invalid")
	}
	sourceRunAt, sourceOK := validTimestamp(value.SourceRunAt)
	occurredAt, occurredOK := validTimestamp(value.OccurredAt)
	if !sourceOK || !occurredOK {
		return fmt.Errorf("stale job disposition DB-clock timestamp is invalid")
	}
	switch value.Decision {
	case StaleJobAdopt:
		if !validOptionalUUID(value.ResultJobID) || *value.ResultJobID == value.SourceJobID ||
			*value.ResultJobID == value.EvidenceRootID || *value.ResultJobID == value.DispositionID ||
			value.ResultPayload == nil || value.ResultRunAt == nil ||
			value.ResultPayload.validate() != nil ||
			!sameStalePayload(value.SourcePayload, *value.ResultPayload) {
			return fmt.Errorf("stale job adoption result is invalid")
		}
		resultRunAt, ok := validTimestamp(*value.ResultRunAt)
		if !ok || !resultRunAt.Equal(maxDispositionTime(sourceRunAt, occurredAt)) {
			return fmt.Errorf("stale job adoption schedule is invalid")
		}
	case StaleJobCancel:
		if value.ResultJobID != nil || value.ResultPayload != nil || value.ResultRunAt != nil {
			return fmt.Errorf("stale job cancellation result is invalid")
		}
	default:
		return fmt.Errorf("stale job decision is invalid")
	}
	return nil
}

func maxDispositionTime(left, right time.Time) time.Time {
	if left.After(right) {
		return left
	}
	return right
}

func (value StaleJobPayloadIdentity) validate() error {
	if value.Revision < 1 || !validDigest(value.SHA256) ||
		!validText(value.Reference, MaximumReferenceBytes) ||
		!validText(value.DedupeKey, MaximumKeyBytes) ||
		(value.EffectKey != nil && !validText(*value.EffectKey, MaximumKeyBytes)) {
		return fmt.Errorf("stale job payload identity is invalid")
	}
	return nil
}

func sameStaleAuthorityIdentity(old Authority, current Authority) bool {
	return old.ScopeKind == current.ScopeKind && old.ScopeID == current.ScopeID &&
		old.GroupID == current.GroupID && old.Domain == current.Domain
}

func sameStalePayload(source StaleJobPayloadIdentity, result StaleJobPayloadIdentity) bool {
	return source.Revision == result.Revision && source.SHA256 == result.SHA256 &&
		((source.EffectKey == nil && result.EffectKey == nil) ||
			(source.EffectKey != nil && result.EffectKey != nil &&
				*source.EffectKey == *result.EffectKey))
}

func cloneStalePayload(value StaleJobPayloadIdentity) StaleJobPayloadIdentity {
	value.EffectKey = cloneOptional(value.EffectKey)
	return value
}

func (value StaleJobDisposition) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire StaleJobDisposition
	return json.Marshal(wire(value))
}

func DecodeStaleJobDisposition(document []byte) (StaleJobDisposition, error) {
	var value StaleJobDisposition
	if err := decodeStrict(document, &value); err != nil {
		return StaleJobDisposition{}, err
	}
	if err := value.Validate(); err != nil {
		return StaleJobDisposition{}, err
	}
	return NewStaleJobDisposition(value)
}
