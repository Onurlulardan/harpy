package systemtransfer

import (
	"context"
	"io"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	"github.com/google/uuid"
)

type ExportPlaintextSpoolIdentity struct {
	TenantID          uuid.UUID
	ExportID          uuid.UUID
	AttemptGeneration int64
}

func (identity ExportPlaintextSpoolIdentity) Validate() error {
	if identity.TenantID == uuid.Nil || !exportV7(identity.ExportID) ||
		identity.AttemptGeneration < 1 {
		return ErrExportWorkerInvalid
	}
	return nil
}

// ExportPlaintextSpool is node-local, bounded storage whose writer must be
// closed before a reader is opened. Discard removes the plaintext evidence.
type ExportPlaintextSpool interface {
	OpenWriter() (io.WriteCloser, error)
	OpenReader() (io.ReadCloser, error)
	Discard() error
}

type ExportPlaintextSpoolAllocator interface {
	ReserveExportPlaintext(
		context.Context,
		ExportPlaintextSpoolIdentity,
		int64,
	) (ExportPlaintextSpool, error)
}

func exportPlaintextMaximumBytes() int64 {
	return tabulartransfer.MaximumDecompressedBytes
}
