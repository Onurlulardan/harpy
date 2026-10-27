package systemtransfer

type ExportWorker struct {
	clock      ExportClock
	loader     ExportAttemptLoader
	lease      ExportJobLease
	scopes     FreshExportScopes
	authorizer ExportCheckpointAuthorizer
	stores     ExportWorkerStoreBinder
	snapshots  ExportSnapshotProvider
	identities ExportIdentitySource
	spools     ExportPlaintextSpoolAllocator
	artifacts  ExportArtifactSealer
}

func NewSystemExportWorker(
	clock ExportClock,
	loader ExportAttemptLoader,
	lease ExportJobLease,
	scopes FreshExportScopes,
	authorizer ExportCheckpointAuthorizer,
	stores ExportWorkerStoreBinder,
	snapshots ExportSnapshotProvider,
	identities ExportIdentitySource,
	spools ExportPlaintextSpoolAllocator,
	artifacts ExportArtifactSealer,
) (*ExportWorker, error) {
	dependencies := []any{clock, loader, lease, scopes, authorizer, stores,
		snapshots, identities, spools, artifacts}
	for _, dependency := range dependencies {
		if nilDependency(dependency) {
			return nil, ErrExportWorkerInvalid
		}
	}
	return &ExportWorker{clock: clock, loader: loader, lease: lease,
		scopes: scopes, authorizer: authorizer, stores: stores,
		snapshots: snapshots, identities: identities, spools: spools,
		artifacts: artifacts}, nil
}
