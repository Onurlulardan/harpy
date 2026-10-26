package iam

import (
	"context"
	"errors"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type systemTransferImportWorker interface {
	ResumeAttempt(context.Context, systemtransfer.AttemptClaim) (
		systemtransfer.Attempt, error)
}

type systemTransferImportTerminal interface {
	Complete(context.Context, asyncpostgres.JobClaim, systemtransfer.Attempt) error
	Fail(context.Context, asyncpostgres.JobClaim, systemtransfer.Attempt,
		systemtransfer.WorkerFailure) error
}

type systemTransferImportJobProcessor struct {
	policy   jobsv1.KindPolicy
	worker   systemTransferImportWorker
	terminal systemTransferImportTerminal
}

func newSystemTransferImportJobProcessor(policy jobsv1.KindPolicy,
	worker systemTransferImportWorker,
	terminal systemTransferImportTerminal,
) (*systemTransferImportJobProcessor, error) {
	if ValidateSystemImportJobPolicy(policy) != nil ||
		nilIdentityDependency(worker) || nilIdentityDependency(terminal) {
		return nil, ErrSystemTransferJobPolicy
	}
	return &systemTransferImportJobProcessor{
		policy: policy, worker: worker, terminal: terminal,
	}, nil
}

func (processor *systemTransferImportJobProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

func (processor *systemTransferImportJobProcessor) Process(ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	if ctx == nil || processor == nil || nilIdentityDependency(processor.worker) ||
		nilIdentityDependency(processor.terminal) {
		return ErrSystemTransferJobPolicy
	}
	decoded, err := decodeSystemTransferImportClaim(claim, processor.policy)
	if err != nil {
		return err
	}
	attempt, err := processor.worker.ResumeAttempt(ctx, decoded)
	if err == nil {
		err = processor.terminal.Complete(ctx, claim, attempt)
		if err == nil {
			return nil
		}
		err = systemtransfer.ClassifyWorkerFailure(err, attempt.ActorID)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, systemtransfer.ErrWorkerLeaseLost) ||
		errors.Is(err, asyncpostgres.ErrJobLeaseLost) {
		return asyncpostgres.ErrJobLeaseLost
	}
	var failure *systemtransfer.WorkerFailure
	if !errors.As(err, &failure) || failure.ActorID == [16]byte{} ||
		attempt.Validate() != nil || !attemptMatchesImportClaim(attempt, decoded) {
		return fmt.Errorf("execute IAM system import: %w", err)
	}
	return processor.terminal.Fail(ctx, claim, attempt, *failure)
}

func attemptMatchesImportClaim(attempt systemtransfer.Attempt,
	claim systemtransfer.AttemptClaim,
) bool {
	return attempt.TenantID == claim.TenantID && attempt.ID == claim.ImportID &&
		attempt.JobID == claim.JobID && attempt.AttemptID == claim.AttemptID &&
		attempt.LeaseOwnerID == claim.LeaseOwnerID &&
		attempt.LeaseGeneration == claim.LeaseGeneration &&
		attempt.Resource == claim.Resource
}
