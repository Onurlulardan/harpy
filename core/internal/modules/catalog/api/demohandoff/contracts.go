package demohandoff

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const Operation = "catalog.operation.resume"

type Request struct {
	ProvisionReceiptID [16]byte
	ComponentID        [16]byte
	PackageID          [16]byte
	ReleaseID          [16]byte
	FormsRunID         [16]byte
	RequestSHA256      [32]byte
	FormCount          int
	RowCount           int
	ReadbackSHA256     [32]byte
	CompletedAt        time.Time
}

type Receipt struct {
	HandoffID          [16]byte
	ProvisionReceiptID [16]byte
	ComponentID        [16]byte
	PackageID          [16]byte
	ReleaseID          [16]byte
	FormsRunID         [16]byte
	RequestSHA256      [32]byte
	FormCount          int
	RowCount           int
	ReadbackSHA256     [32]byte
	CompletedAt        time.Time
}

type Recorder interface {
	RecordDemoHandoff(
		context.Context,
		platformtx.TenantScope,
		Request,
	) (Receipt, error)
}
