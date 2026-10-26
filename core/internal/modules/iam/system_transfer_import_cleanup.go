package iam

import (
	"fmt"

	cleanupstore "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfercleanup"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/google/uuid"
)

type SystemTransferImportCleanupRuntime = systemtransfer.SourceCleanupRunner
type SystemTransferImportCleanupClock = systemtransfer.SourceCleanupClock
type SystemTransferImportCleanupIDs = systemtransfer.SourceCleanupIDs
type SystemTransferImportCleanupJitter = systemtransfer.SourceCleanupJitter
type SystemTransferImportCleanupID = uuid.UUID

var BindSystemTransferImportCleanupStore = cleanupstore.Bind

func NewSystemTransferImportCleanupRuntime(
	store systemtransfer.SourceCleanupStore,
	objects systemtransfer.SourceObjectDeleter,
	clock SystemTransferImportCleanupClock,
	ids SystemTransferImportCleanupIDs,
	jitter SystemTransferImportCleanupJitter,
	gate *admission.Gate,
) (*SystemTransferImportCleanupRuntime, error) {
	runtime, err := systemtransfer.NewSourceCleanupRunner(
		store, objects, clock, ids, jitter, gate,
	)
	if err != nil {
		return nil, fmt.Errorf("build IAM system import cleanup runtime: %w", err)
	}
	return runtime, nil
}
