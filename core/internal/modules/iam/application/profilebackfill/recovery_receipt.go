package profilebackfill

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

const (
	RecoveryRunning RecoveryState = "running"
	RecoveryFenced  RecoveryState = "fenced"
)

type RecoveryReceipt struct {
	RequestSHA256      [sha256.Size]byte
	PlanID             uuid.UUID
	RecoveryOfPlanID   uuid.UUID
	RecoveryRootPlanID uuid.UUID
	RecoverySequence   int64
	Revision           int64
	State              RecoveryState
	AuthorizationClock int64
}

func (receipt RecoveryReceipt) Validate() error {
	if receipt.RequestSHA256 == [sha256.Size]byte{} ||
		!stableID(receipt.PlanID) || !stableID(receipt.RecoveryOfPlanID) ||
		!stableID(receipt.RecoveryRootPlanID) || receipt.RecoverySequence < 1 ||
		receipt.PlanID == receipt.RecoveryOfPlanID ||
		receipt.PlanID == receipt.RecoveryRootPlanID || receipt.Revision < 1 ||
		receipt.AuthorizationClock < 1 || !receipt.State.Valid() {
		return ErrInvalid
	}
	return nil
}

func (state RecoveryState) Valid() bool {
	return state == RecoveryPlanned || state == RecoveryRunning ||
		state == RecoveryActivated || state == RecoveryFenced
}
