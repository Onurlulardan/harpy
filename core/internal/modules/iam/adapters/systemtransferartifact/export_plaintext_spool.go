package systemtransferartifact

import (
	"context"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (runtime *Runtime) ReserveExportPlaintext(
	ctx context.Context,
	identity systemtransfer.ExportPlaintextSpoolIdentity,
	maximumBytes int64,
) (systemtransfer.ExportPlaintextSpool, error) {
	if runtime == nil || !runtimeInputContext(ctx) || identity.Validate() != nil ||
		maximumBytes != tabulartransfer.MaximumDecompressedBytes {
		return nil, errInvalid
	}
	return runtime.acquireSpool(identity.ExportID, identity.AttemptGeneration,
		identity.TenantID, false)
}

var _ systemtransfer.ExportPlaintextSpoolAllocator = (*Runtime)(nil)
