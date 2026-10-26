package installationdelivery

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

type State string

const (
	StateReady          State = "ready_to_display"
	StateClaimed        State = "claimed"
	StateAcknowledged   State = "acknowledged"
	StateOutcomeUnknown State = "outcome_unknown"
	StateSecretDeleted  State = "secret_deleted"
	StateCompleted      State = "completed"
)

type Delivery struct {
	OperationID, TenantID, FirstOwnerUserID          uuid.UUID
	CurrentAttemptID, CurrentCredentialID            uuid.UUID
	CurrentCredentialVersion, Revision               int64
	State                                            State
	DatabaseNow                                      time.Time
	ClaimAdapterID                                   string
	ClaimLeaseUntil, AcknowledgedAt, SecretDeletedAt *time.Time
	CompletedAt                                      *time.Time
}

func (delivery Delivery) Valid() bool {
	if delivery.OperationID == uuid.Nil || delivery.TenantID == uuid.Nil ||
		delivery.FirstOwnerUserID == uuid.Nil || delivery.CurrentAttemptID == uuid.Nil ||
		delivery.CurrentCredentialID == uuid.Nil || delivery.CurrentCredentialVersion < 1 ||
		delivery.Revision < 1 || !utc(delivery.DatabaseNow) {
		return false
	}
	switch delivery.State {
	case StateReady:
		return delivery.ClaimAdapterID == "" && delivery.ClaimLeaseUntil == nil
	case StateClaimed, StateOutcomeUnknown:
		return validAdapter(delivery.ClaimAdapterID) && validTime(delivery.ClaimLeaseUntil)
	case StateAcknowledged:
		return validAdapter(delivery.ClaimAdapterID) && validTime(delivery.AcknowledgedAt)
	case StateSecretDeleted:
		return validTime(delivery.AcknowledgedAt) && validTime(delivery.SecretDeletedAt)
	case StateCompleted:
		return validTime(delivery.AcknowledgedAt) && validTime(delivery.SecretDeletedAt) &&
			validTime(delivery.CompletedAt)
	default:
		return false
	}
}

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validDigest(value string) bool   { return digestPattern.MatchString(value) }
func validAdapter(value string) bool  { return len(value) >= 1 && len(value) <= 128 }
func utc(value time.Time) bool        { return !value.IsZero() && value.Location() == time.UTC }
func validTime(value *time.Time) bool { return value != nil && utc(*value) }
