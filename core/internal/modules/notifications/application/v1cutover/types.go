package v1cutover

import (
	"time"

	"github.com/google/uuid"
)

type Origin string

const (
	OriginFresh     Origin = "fresh"
	OriginV1Adopted Origin = "v1_adopted"
)

type Authority struct {
	Writer, State          string
	Generation, RowVersion int64
}

func (value Authority) Valid() bool {
	return (value.Writer == "v1" || value.Writer == "v2") &&
		(value.State == "writable" || value.State == "quiescing" || value.State == "fenced") &&
		value.Generation > 0 && value.RowVersion > 0
}

type Inventory struct {
	RunID                          uuid.UUID
	SourceQueueDepth               int64
	FutureDeliveryCount            int64
	RunningDeliveryCount           int64
	UnresolvedProviderReceiptCount int64
	LegacyConnectionCount          int64
	SourceEvidenceDigest           [32]byte
	ObservedAt                     time.Time
}

func (value Inventory) Validate() error {
	if value.RunID == uuid.Nil || value.SourceQueueDepth < 0 ||
		value.FutureDeliveryCount < 0 || value.RunningDeliveryCount < 0 ||
		value.UnresolvedProviderReceiptCount < 0 || value.LegacyConnectionCount < 0 ||
		value.SourceEvidenceDigest == [32]byte{} || value.ObservedAt.IsZero() {
		return ErrInvalid
	}
	return nil
}

func (value Inventory) Zero() bool {
	return value.SourceQueueDepth == 0 && value.FutureDeliveryCount == 0 &&
		value.RunningDeliveryCount == 0 && value.UnresolvedProviderReceiptCount == 0 &&
		value.LegacyConnectionCount == 0
}

type Evidence struct {
	ID        uuid.UUID
	Inventory Inventory
	ExpiresAt time.Time
}
