package iam

import (
	"context"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type systemTransferExportReconciliationWorker interface {
	ProcessExportReconciliation(
		context.Context,
		systemtransfer.ExportReconciliationClaim,
	) error
}

type systemTransferExportReconcileJobProcessor struct {
	policy jobsv1.KindPolicy
	worker systemTransferExportReconciliationWorker
}

func newSystemTransferExportReconcileJobProcessor(
	policy jobsv1.KindPolicy,
	worker systemTransferExportReconciliationWorker,
) (*systemTransferExportReconcileJobProcessor, error) {
	if ValidateSystemExportReconcileJobPolicy(policy) != nil ||
		nilIdentityDependency(worker) {
		return nil, ErrSystemTransferJobPolicy
	}
	return &systemTransferExportReconcileJobProcessor{policy: policy, worker: worker}, nil
}

func (processor *systemTransferExportReconcileJobProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

func (processor *systemTransferExportReconcileJobProcessor) Process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	if processor == nil || ctx == nil || nilIdentityDependency(processor.worker) {
		return ErrSystemTransferJobPolicy
	}
	decoded, err := decodeSystemTransferExportReconcileClaim(claim, processor.policy)
	if err != nil {
		return err
	}
	return processor.worker.ProcessExportReconciliation(ctx, decoded)
}
