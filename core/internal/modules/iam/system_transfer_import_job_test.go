package iam

import (
	"context"
	"testing"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type importWorkerStub struct{}

func (*importWorkerStub) ResumeAttempt(context.Context, systemtransfer.AttemptClaim) (
	systemtransfer.Attempt, error,
) {
	return systemtransfer.Attempt{}, nil
}

type importTerminalStub struct{}

func (*importTerminalStub) Complete(context.Context, asyncpostgres.JobClaim,
	systemtransfer.Attempt,
) error {
	return nil
}

func (*importTerminalStub) Fail(context.Context, asyncpostgres.JobClaim,
	systemtransfer.Attempt, systemtransfer.WorkerFailure,
) error {
	return nil
}

func TestSystemTransferImportProcessorAcceptsOnlyCanonicalPolicy(t *testing.T) {
	processor, err := newSystemTransferImportJobProcessor(
		SystemImportJobPolicy(), &importWorkerStub{}, &importTerminalStub{},
	)
	if err != nil || processor.Policy().Kind != SystemImportJobKind {
		t.Fatalf("processor=%#v err=%v", processor, err)
	}
	policy := SystemImportJobPolicy()
	policy.MaximumAttempts--
	if _, err = newSystemTransferImportJobProcessor(
		policy, &importWorkerStub{}, &importTerminalStub{},
	); err == nil {
		t.Fatal("non-canonical retry budget was accepted")
	}
}
