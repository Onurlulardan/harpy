package iam

import (
	"fmt"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func NewSystemTransferExportReconciliationProcessor(
	release platformasync.ReleaseCatalogs,
	store systemtransfer.ExportReconciliationStore,
	artifacts SystemTransferArtifacts,
) (jobexecution.Processor, error) {
	if nilIdentityDependency(store) || nilIdentityDependency(artifacts) {
		return nil, systemtransfer.ErrExportReconciliationInvalid
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read IAM export reconciliation job catalog: %w", err)
	}
	policy, err := jobs.Resolve(SystemExportReconcileJobKind, 1)
	if err != nil || ValidateSystemExportReconcileJobPolicy(policy) != nil {
		return nil, ErrSystemTransferJobPolicy
	}
	backoff := newSystemTransferExportReconciliationBackoff()
	worker, err := systemtransfer.NewSystemExportReconciliationProcessor(
		store, artifacts, systemTransferArtifactClock{}, backoff,
	)
	if err != nil {
		return nil, fmt.Errorf("build IAM export reconciliation application: %w", err)
	}
	processor, err := newSystemTransferExportReconcileJobProcessor(policy, worker)
	if err != nil {
		return nil, err
	}
	return processor, nil
}
