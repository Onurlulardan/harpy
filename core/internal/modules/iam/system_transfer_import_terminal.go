package iam

import (
	"context"
	"fmt"

	systemtransferpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfer"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferImportFinisher interface {
	CommitTerminal(context.Context, platformtx.TenantScope, publicresource.ActorEvidence,
		systemtransfer.Attempt, bool) (systemtransferpostgres.WorkerTerminal, error)
}

type systemTransferImportJobTerminalRuntime struct {
	scopes     systemtransfer.FreshBatchScopes
	authorizer systemtransfer.BatchAuthorizer
	finisher   systemTransferImportFinisher
	catalogs   platformasync.WriteCatalogs
	jitter     platformasync.JitterSource
}

func newSystemTransferImportJobTerminal(scopes systemtransfer.FreshBatchScopes,
	authorizer systemtransfer.BatchAuthorizer,
	finisher systemTransferImportFinisher,
	catalogs platformasync.WriteCatalogs,
) (*systemTransferImportJobTerminalRuntime, error) {
	if nilIdentityDependency(scopes) || nilIdentityDependency(authorizer) ||
		nilIdentityDependency(finisher) {
		return nil, systemtransfer.ErrDependencies
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, fmt.Errorf("read IAM import terminal catalogs: %w", err)
	}
	return &systemTransferImportJobTerminalRuntime{
		scopes: scopes, authorizer: authorizer, finisher: finisher,
		catalogs: catalogs, jitter: systemTransferImportJitter{},
	}, nil
}

func (terminal *systemTransferImportJobTerminalRuntime) Complete(ctx context.Context,
	claim asyncpostgres.JobClaim,
	attempt systemtransfer.Attempt,
) error {
	if terminal == nil || attempt.Validate() != nil {
		return systemtransfer.ErrInvalidAttempt
	}
	return terminal.scopes.InFreshImportBatch(ctx, attempt,
		func(scoped context.Context, batch systemtransfer.BatchScope) error {
			evidence, err := terminal.authorizer.AuthorizeImportBatch(
				scoped, batch.IAMScope(), attempt,
			)
			if err != nil {
				return err
			}
			if _, err = terminal.finisher.CommitTerminal(
				scoped, batch.IAMScope(), evidence, attempt, true,
			); err != nil {
				return err
			}
			writer, err := terminal.writer(batch.IAMScope())
			if err != nil {
				return err
			}
			_, err = writer.CompleteDatabaseJob(scoped, claim)
			return err
		})
}

func (terminal *systemTransferImportJobTerminalRuntime) writer(
	scope platformtx.TenantScope,
) (asyncpostgres.Writer, error) {
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil {
		return nil, err
	}
	return asyncpostgres.Bind(scope, authority, terminal.catalogs)
}

var _ systemTransferImportTerminal = (*systemTransferImportJobTerminalRuntime)(nil)
var _ systemTransferImportFinisher = (*systemtransferpostgres.WorkerStore)(nil)
