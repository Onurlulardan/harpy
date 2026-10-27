package iam

import (
	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	transferartifact "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/systemtransferartifact"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

const (
	SystemTransferImportObjectPrefix     = transferartifact.ImportObjectPrefix
	SystemTransferImportContract         = transferartifact.ImportObjectContract
	SystemTransferExportObjectPrefix     = transferartifact.ExportObjectPrefix
	SystemTransferExportContract         = transferartifact.ExportObjectContract
	SystemTransferMaximumCiphertextBytes = int64(fcartv1.MaxCiphertextBytes)
)

type SystemTransferArtifactObjectStore = transferartifact.ObjectStore
type SystemTransferArtifactSpoolManager = transferartifact.SpoolManager

type SystemTransferArtifactRuntime = systemtransfer.ArtifactRuntime
type SystemTransferArtifacts interface {
	SystemTransferArtifactRuntime
	systemtransfer.ExactImportSourceOpener
	systemtransfer.SourceObjectDeleter
	systemtransfer.ExportArtifactSealer
	systemtransfer.ExportPlaintextSpoolAllocator
	systemtransfer.ExportArtifactOutcomeReconciler
}
