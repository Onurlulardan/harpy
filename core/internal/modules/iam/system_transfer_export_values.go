package iam

import (
	"time"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/google/uuid"
)

type systemTransferExportValues struct{}

func (systemTransferExportValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func (systemTransferExportValues) NewExportIdentity(
	at time.Time,
) (uuid.UUID, error) {
	return randomid.NewUUIDv7(at)
}

var _ systemtransfer.ExportClock = systemTransferExportValues{}
var _ systemtransfer.ExportIdentitySource = systemTransferExportValues{}
