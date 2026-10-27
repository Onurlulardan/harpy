package billing

import (
	"errors"

	"github.com/google/uuid"
)

var ErrDocumentsUsageDenied = errors.New("documents storage entitlement denied the operation")

type DocumentsUsageReservation struct {
	ID      uuid.UUID
	Allowed bool
	Replay  bool
}
