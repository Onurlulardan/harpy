package jobsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type RetentionRelease struct {
	ContractVersion        string    `json:"contract_version"`
	JobID                  string    `json:"job_id"`
	TenantID               *string   `json:"tenant_id"`
	Authority              Authority `json:"authority"`
	ReleaseRevision        int64     `json:"retention_release_revision"`
	RetainUntil            string    `json:"retain_until"`
	LiveDependencyCount    int       `json:"live_dependency_count"`
	RetryChildCount        int       `json:"retry_child_count"`
	UnfinishedAttemptCount int       `json:"unfinished_attempt_count"`
	LiveEffectReceiptCount int       `json:"live_effect_receipt_count"`
}

func NewRetentionRelease(value RetentionRelease) (RetentionRelease, error) {
	value.ContractVersion = RetentionReleaseContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	if err := value.Validate(); err != nil {
		return RetentionRelease{}, err
	}
	return value, nil
}

func (value RetentionRelease) Validate() error {
	if value.ContractVersion != RetentionReleaseContractVersion ||
		!validUUID(value.JobID) || value.ReleaseRevision < 1 {
		return fmt.Errorf("job retention release identity is invalid")
	}
	if err := value.Authority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if value.Authority.WriterOrigin != WriterV2 {
		return fmt.Errorf("job retention release requires V2 authority")
	}
	if _, ok := validTimestamp(value.RetainUntil); !ok {
		return fmt.Errorf("job retention release retain_until is invalid")
	}
	if value.LiveDependencyCount != 0 || value.RetryChildCount != 0 ||
		value.UnfinishedAttemptCount != 0 || value.LiveEffectReceiptCount != 0 {
		return fmt.Errorf("job retention release requires zero live references")
	}
	return nil
}

func (value RetentionRelease) CanonicalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire RetentionRelease
	document, err := json.Marshal(wire(value))
	if err != nil {
		return nil, fmt.Errorf("marshal job retention release: %w", err)
	}
	return canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
}

func (value RetentionRelease) SHA256() (string, error) {
	document, err := value.CanonicalJSON()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:]), nil
}

func (value RetentionRelease) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire RetentionRelease
	return json.Marshal(wire(value))
}

func DecodeRetentionRelease(document []byte) (RetentionRelease, error) {
	var value RetentionRelease
	if err := decodeStrict(document, &value); err != nil {
		return RetentionRelease{}, err
	}
	return NewRetentionRelease(value)
}
