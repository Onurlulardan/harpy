package jobsv1

import (
	"encoding/json"
	"fmt"
)

type KindPolicy struct {
	ContractVersion       string      `json:"contract_version"`
	Kind                  string      `json:"kind"`
	PayloadRevision       int         `json:"payload_revision"`
	OwnerDomain           string      `json:"owner_domain"`
	EffectClass           EffectClass `json:"effect_class"`
	MaximumAttempts       int         `json:"maximum_attempts"`
	RetryMode             RetryMode   `json:"retry_mode"`
	RetryBaseMilliseconds int         `json:"retry_base_milliseconds"`
	RetryCapMilliseconds  int         `json:"retry_cap_milliseconds"`
	LeaseSeconds          int         `json:"lease_seconds"`
	HeartbeatSeconds      int         `json:"heartbeat_seconds"`
	SuccessRetentionDays  int         `json:"success_retention_days"`
	DeadRetentionDays     int         `json:"dead_retention_days"`
	DetailRedactionDays   int         `json:"detail_redaction_days"`
}

func NewKindPolicy(kind string, revision int, domain string, effect EffectClass, attempts int) (KindPolicy, error) {
	value := KindPolicy{
		ContractVersion: KindPolicyContractVersion, Kind: kind, PayloadRevision: revision,
		OwnerDomain: domain, EffectClass: effect, MaximumAttempts: attempts,
		RetryMode: RetryFullJitter, RetryBaseMilliseconds: RetryBaseMilliseconds,
		RetryCapMilliseconds: RetryCapMilliseconds, LeaseSeconds: LeaseSeconds,
		HeartbeatSeconds: HeartbeatSeconds, SuccessRetentionDays: SuccessfulRetentionDays,
		DeadRetentionDays: DeadRetentionDays, DetailRedactionDays: PayloadAndErrorRedactionDays,
	}
	if err := value.Validate(); err != nil {
		return KindPolicy{}, err
	}
	return value, nil
}

func (value KindPolicy) Validate() error {
	if value.ContractVersion != KindPolicyContractVersion || !validKind(value.Kind) ||
		value.PayloadRevision < 1 || !validDomain(value.OwnerDomain) || !value.EffectClass.valid() {
		return fmt.Errorf("job kind policy identity is invalid")
	}
	if value.MaximumAttempts < MinimumMaximumAttempts || value.MaximumAttempts > MaximumMaximumAttempts {
		return fmt.Errorf("maximum attempts must be between 1 and 25")
	}
	if value.RetryMode != RetryFullJitter || value.RetryBaseMilliseconds != RetryBaseMilliseconds ||
		value.RetryCapMilliseconds != RetryCapMilliseconds || value.LeaseSeconds != LeaseSeconds ||
		value.HeartbeatSeconds != HeartbeatSeconds {
		return fmt.Errorf("job execution policy differs from ADR-0011")
	}
	if value.SuccessRetentionDays != SuccessfulRetentionDays || value.DeadRetentionDays != DeadRetentionDays ||
		value.DetailRedactionDays != PayloadAndErrorRedactionDays {
		return fmt.Errorf("job retention policy differs from ADR-0011")
	}
	return nil
}

func (value KindPolicy) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire KindPolicy
	return json.Marshal(wire(value))
}

func DecodeKindPolicy(document []byte) (KindPolicy, error) {
	var value KindPolicy
	if err := decodeStrict(document, &value); err != nil {
		return KindPolicy{}, err
	}
	return value, value.Validate()
}
