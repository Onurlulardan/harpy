package jobsv1

import (
	"encoding/json"
	"fmt"
	"time"
)

type Lease struct {
	ContractVersion  string    `json:"contract_version"`
	JobID            string    `json:"job_id"`
	Kind             string    `json:"kind"`
	PayloadRevision  int       `json:"payload_revision"`
	TenantID         *string   `json:"tenant_id"`
	Authority        Authority `json:"authority"`
	State            State     `json:"state"`
	LeaseOwner       string    `json:"lease_owner"`
	LeaseGeneration  int64     `json:"lease_generation"`
	AttemptNumber    int       `json:"attempt_number"`
	LeasedAt         string    `json:"leased_at"`
	LeaseExpiresAt   string    `json:"lease_expires_at"`
	HeartbeatSeconds int       `json:"heartbeat_seconds"`
	policy           KindPolicy
}

func NewLease(value Lease, catalog Catalog) (Lease, error) {
	value.ContractVersion = LeaseContractVersion
	value.State = StateRunning
	value.HeartbeatSeconds = HeartbeatSeconds
	value.TenantID = cloneOptional(value.TenantID)
	if err := value.Validate(catalog); err != nil {
		return Lease{}, err
	}
	value.policy, _ = catalog.Resolve(value.Kind, value.PayloadRevision)
	return value, nil
}

func (value Lease) validateSyntax() error {
	if value.ContractVersion != LeaseContractVersion || !validUUID(value.JobID) ||
		!validKind(value.Kind) || value.PayloadRevision < 1 || value.State != StateRunning {
		return fmt.Errorf("job lease identity or state is invalid")
	}
	if err := value.Authority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if !validUUID(value.LeaseOwner) || value.LeaseGeneration < 1 || value.AttemptNumber < 1 ||
		int64(value.AttemptNumber) != value.LeaseGeneration {
		return fmt.Errorf("job lease fencing values are invalid")
	}
	leasedAt, leasedOK := validTimestamp(value.LeasedAt)
	expiresAt, expiresOK := validTimestamp(value.LeaseExpiresAt)
	if !leasedOK || !expiresOK || !expiresAt.Equal(leasedAt.Add(LeaseSeconds*time.Second)) {
		return fmt.Errorf("job lease interval must be exactly 60 seconds")
	}
	if value.HeartbeatSeconds != HeartbeatSeconds {
		return fmt.Errorf("job heartbeat interval must be 20 seconds")
	}
	return nil
}

func (value Lease) Validate(catalog Catalog) error {
	if err := value.validateSyntax(); err != nil {
		return err
	}
	policy, err := catalog.Resolve(value.Kind, value.PayloadRevision)
	if err != nil {
		return err
	}
	if value.Authority.Domain != policy.OwnerDomain || value.AttemptNumber > policy.MaximumAttempts {
		return fmt.Errorf("job lease differs from registered policy")
	}
	return nil
}

func (value Lease) MarshalJSON() ([]byte, error) {
	if err := value.Validate(Catalog{policies: []KindPolicy{value.policy}}); err != nil {
		return nil, err
	}
	type wire Lease
	return json.Marshal(wire(value))
}

func DecodeLease(document []byte, catalog Catalog) (Lease, error) {
	var value Lease
	if err := decodeStrict(document, &value); err != nil {
		return Lease{}, err
	}
	if err := value.Validate(catalog); err != nil {
		return Lease{}, err
	}
	return NewLease(value, catalog)
}
