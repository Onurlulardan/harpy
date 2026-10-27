package jobsv1

import (
	"encoding/json"
	"fmt"
)

type Enqueue struct {
	ContractVersion      string    `json:"contract_version"`
	JobID                string    `json:"job_id"`
	Kind                 string    `json:"kind"`
	PayloadRevision      int       `json:"payload_revision"`
	TenantID             *string   `json:"tenant_id"`
	Authority            Authority `json:"authority"`
	PayloadReference     string    `json:"payload_reference"`
	PayloadSHA256        string    `json:"payload_sha256"`
	RunAt                string    `json:"run_at"`
	Priority             int32     `json:"priority"`
	DedupeKey            string    `json:"dedupe_key"`
	EffectIdempotencyKey *string   `json:"effect_idempotency_key"`
	MaximumAttempts      int       `json:"maximum_attempts"`
	RetryOfJobID         *string   `json:"retry_of_job_id"`
	policy               KindPolicy
}

func NewEnqueue(value Enqueue, catalog Catalog) (Enqueue, error) {
	value.ContractVersion = EnqueueContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	value.EffectIdempotencyKey = cloneOptional(value.EffectIdempotencyKey)
	value.RetryOfJobID = cloneOptional(value.RetryOfJobID)
	if err := value.Validate(catalog); err != nil {
		return Enqueue{}, err
	}
	value.policy, _ = catalog.Resolve(value.Kind, value.PayloadRevision)
	return value, nil
}

func (value Enqueue) validateSyntax() error {
	if value.ContractVersion != EnqueueContractVersion || !validUUID(value.JobID) ||
		!validKind(value.Kind) || value.PayloadRevision < 1 {
		return fmt.Errorf("job enqueue identity is invalid")
	}
	if err := value.Authority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if !validText(value.PayloadReference, MaximumReferenceBytes) || !validDigest(value.PayloadSHA256) {
		return fmt.Errorf("job payload reference is invalid")
	}
	if _, ok := validTimestamp(value.RunAt); !ok {
		return fmt.Errorf("job run_at is invalid")
	}
	if !validText(value.DedupeKey, MaximumKeyBytes) {
		return fmt.Errorf("job dedupe key is invalid")
	}
	if value.EffectIdempotencyKey != nil && !validText(*value.EffectIdempotencyKey, MaximumKeyBytes) {
		return fmt.Errorf("job effect idempotency key is invalid")
	}
	if value.RetryOfJobID != nil && (!validUUID(*value.RetryOfJobID) || *value.RetryOfJobID == value.JobID) {
		return fmt.Errorf("retry source job ID is invalid")
	}
	return nil
}

func (value Enqueue) Validate(catalog Catalog) error {
	if err := value.validateSyntax(); err != nil {
		return err
	}
	policy, err := catalog.Resolve(value.Kind, value.PayloadRevision)
	if err != nil {
		return err
	}
	if value.Authority.Domain != policy.OwnerDomain || value.MaximumAttempts != policy.MaximumAttempts {
		return fmt.Errorf("job enqueue differs from registered policy")
	}
	if policy.EffectClass == EffectExternal && value.EffectIdempotencyKey == nil ||
		policy.EffectClass == EffectDatabaseTransactional && value.EffectIdempotencyKey != nil {
		return fmt.Errorf("job effect idempotency key differs from effect policy")
	}
	return nil
}

func (value Enqueue) MarshalJSON() ([]byte, error) {
	if err := value.Validate(Catalog{policies: []KindPolicy{value.policy}}); err != nil {
		return nil, err
	}
	type wire Enqueue
	return json.Marshal(wire(value))
}

func DecodeEnqueue(document []byte, catalog Catalog) (Enqueue, error) {
	var value Enqueue
	if err := decodeStrict(document, &value); err != nil {
		return Enqueue{}, err
	}
	if err := value.Validate(catalog); err != nil {
		return Enqueue{}, err
	}
	return NewEnqueue(value, catalog)
}
