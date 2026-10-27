package systemtransfer

import (
	"time"

	"github.com/google/uuid"
)

type ExportIdentitySource interface {
	NewExportIdentity(time.Time) (uuid.UUID, error)
}
