package jobsv1

import (
	"encoding/json"
	"fmt"
	"time"
)

type AttemptOutcome struct {
	ContractVersion     string      `json:"contract_version"`
	AttemptID           string      `json:"attempt_id"`
	JobID               string      `json:"job_id"`
	Kind                string      `json:"kind"`
	PayloadRevision     int         `json:"payload_revision"`
	AttemptNumber       int         `json:"attempt_number"`
	LeaseOwner          string      `json:"lease_owner"`
	LeaseGeneration     int64       `json:"lease_generation"`
	StartedAt           string      `json:"started_at"`
	EffectStartedAt     *string     `json:"effect_started_at"`
	CompletedAt         string      `json:"completed_at"`
	ResultState         State       `json:"result_state"`
	ErrorClass          *ErrorClass `json:"error_class"`
	ErrorCode           *string     `json:"error_code"`
	RedactedErrorSHA256 *string     `json:"redacted_error_sha256"`
	EffectReceiptID     *string     `json:"effect_receipt_id"`
	RetryRunAt          *string     `json:"retry_run_at"`
	policy              KindPolicy
}

func NewAttemptOutcome(value AttemptOutcome, catalog Catalog) (AttemptOutcome, error) {
	value.ContractVersion = AttemptOutcomeContractVersion
	value.EffectStartedAt = cloneOptional(value.EffectStartedAt)
	value.ErrorClass = cloneErrorClass(value.ErrorClass)
	value.ErrorCode = cloneOptional(value.ErrorCode)
	value.RedactedErrorSHA256 = cloneOptional(value.RedactedErrorSHA256)
	value.EffectReceiptID = cloneOptional(value.EffectReceiptID)
	value.RetryRunAt = cloneOptional(value.RetryRunAt)
	if err := value.Validate(catalog); err != nil {
		return AttemptOutcome{}, err
	}
	value.policy, _ = catalog.Resolve(value.Kind, value.PayloadRevision)
	return value, nil
}

func (value AttemptOutcome) Validate(catalog Catalog) error {
	if value.ContractVersion != AttemptOutcomeContractVersion || !validUUID(value.AttemptID) ||
		!validUUID(value.JobID) || !validKind(value.Kind) || value.PayloadRevision < 1 ||
		value.AttemptNumber < 1 || int64(value.AttemptNumber) != value.LeaseGeneration ||
		!validUUID(value.LeaseOwner) {
		return fmt.Errorf("job attempt identity or fence is invalid")
	}
	started, startedOK := validTimestamp(value.StartedAt)
	completed, completedOK := validTimestamp(value.CompletedAt)
	if !startedOK || !completedOK || completed.Before(started) {
		return fmt.Errorf("job attempt time range is invalid")
	}
	policy, err := catalog.Resolve(value.Kind, value.PayloadRevision)
	if err != nil || value.AttemptNumber > policy.MaximumAttempts {
		return fmt.Errorf("job attempt is not allowed by policy")
	}
	if value.ErrorClass != nil && !value.ErrorClass.valid() {
		return fmt.Errorf("job error class is invalid")
	}
	if value.EffectReceiptID != nil && !validUUID(*value.EffectReceiptID) {
		return fmt.Errorf("job effect receipt ID is invalid")
	}
	if err := value.validateEffectBoundary(policy, started, completed); err != nil {
		return err
	}
	return value.validateOutcome(policy)
}

func (value AttemptOutcome) validateEffectBoundary(policy KindPolicy, started, completed time.Time) error {
	if policy.EffectClass == EffectDatabaseTransactional {
		if value.EffectStartedAt != nil || value.EffectReceiptID != nil {
			return fmt.Errorf("transactional attempt cannot carry external effect boundary")
		}
		return nil
	}
	if value.EffectStartedAt != nil {
		effectStarted, ok := validTimestamp(*value.EffectStartedAt)
		if !ok || effectStarted.Before(started) || effectStarted.After(completed) {
			return fmt.Errorf("external effect start timestamp is invalid")
		}
	}
	if value.ResultState == StateSucceeded &&
		(value.EffectStartedAt == nil || value.EffectReceiptID == nil) {
		return fmt.Errorf("external success requires effect start and receipt")
	}
	if value.ResultState == StateDead && value.ErrorCode != nil && *value.ErrorCode == UnknownOutcomeCode {
		if value.EffectStartedAt == nil || value.EffectReceiptID != nil {
			return fmt.Errorf("unknown outcome requires effect start without receipt")
		}
		return nil
	}
	if (value.EffectStartedAt == nil) != (value.EffectReceiptID == nil) {
		return fmt.Errorf("known external result requires both effect start and receipt, or neither")
	}
	return nil
}

func (value AttemptOutcome) validateOutcome(policy KindPolicy) error {
	switch value.ResultState {
	case StateSucceeded:
		if value.ErrorClass != nil || value.ErrorCode != nil || value.RedactedErrorSHA256 != nil ||
			value.RetryRunAt != nil {
			return fmt.Errorf("successful attempt has error or retry fields")
		}
	case StateQueued:
		return value.validateRetry(policy)
	case StateDead:
		if value.ErrorCode != nil && *value.ErrorCode == UnknownOutcomeCode {
			if policy.EffectClass == EffectExternal && value.ErrorClass == nil &&
				value.RedactedErrorSHA256 == nil && value.RetryRunAt == nil {
				return nil
			}
			return fmt.Errorf("unknown external outcome fields are invalid")
		}
		if value.ErrorClass == nil || value.ErrorCode == nil || value.RedactedErrorSHA256 == nil ||
			!validDigest(*value.RedactedErrorSHA256) || value.RetryRunAt != nil ||
			!validMachineCode(*value.ErrorCode) {
			return fmt.Errorf("dead attempt fields are invalid")
		}
	case StateCancelled:
		if value.ErrorClass == nil || *value.ErrorClass != ErrorCancelled || value.ErrorCode == nil ||
			value.RedactedErrorSHA256 == nil || !validDigest(*value.RedactedErrorSHA256) ||
			value.RetryRunAt != nil || !validMachineCode(*value.ErrorCode) {
			return fmt.Errorf("cancelled attempt fields are invalid")
		}
	default:
		return fmt.Errorf("attempt result state is invalid")
	}
	return nil
}

func (value AttemptOutcome) validateRetry(policy KindPolicy) error {
	if value.RetryRunAt == nil || value.ErrorClass == nil || value.ErrorCode == nil ||
		value.RedactedErrorSHA256 == nil || !validDigest(*value.RedactedErrorSHA256) ||
		(*value.ErrorClass != ErrorTransient && *value.ErrorClass != ErrorConflict) ||
		!validMachineCode(*value.ErrorCode) {
		return fmt.Errorf("retry attempt fields are invalid")
	}
	retryAt, ok := validTimestamp(*value.RetryRunAt)
	completed, _ := validTimestamp(value.CompletedAt)
	if !ok || retryAt.Before(completed) || value.AttemptNumber >= policy.MaximumAttempts ||
		retryAt.Sub(completed) > maximumRetryDelay(value.AttemptNumber) {
		return fmt.Errorf("retry run_at is invalid")
	}
	return nil
}

func (value AttemptOutcome) MarshalJSON() ([]byte, error) {
	if err := value.Validate(Catalog{policies: []KindPolicy{value.policy}}); err != nil {
		return nil, err
	}
	type wire AttemptOutcome
	return json.Marshal(wire(value))
}

func DecodeAttemptOutcome(document []byte, catalog Catalog) (AttemptOutcome, error) {
	var value AttemptOutcome
	if err := decodeStrict(document, &value); err != nil {
		return AttemptOutcome{}, err
	}
	if err := value.Validate(catalog); err != nil {
		return AttemptOutcome{}, err
	}
	return NewAttemptOutcome(value, catalog)
}
