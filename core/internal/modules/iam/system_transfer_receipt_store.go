package iam

import (
	"context"

	systemtransferpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfer"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferReceiptPersistence interface {
	LoadReceipts(context.Context, platformtx.TenantScope,
		publicresource.ActorEvidence, systemtransfer.Attempt,
		[]int64) (map[int64]systemtransfer.RowReceipt, error)
	CommitReceipts(context.Context, platformtx.TenantScope,
		publicresource.ActorEvidence, systemtransfer.Attempt,
		[]systemtransfer.RowReceipt, int64) error
}

type systemTransferReceiptStore struct {
	scopes      systemtransfer.FreshBatchScopes
	authorizer  systemtransfer.BatchAuthorizer
	persistence systemTransferReceiptPersistence
}

func newSystemTransferReceiptStore(scopes systemtransfer.FreshBatchScopes,
	authorizer systemtransfer.BatchAuthorizer,
	persistence systemTransferReceiptPersistence,
) (*systemTransferReceiptStore, error) {
	if nilIdentityDependency(scopes) || nilIdentityDependency(authorizer) ||
		nilIdentityDependency(persistence) {
		return nil, systemtransfer.ErrDependencies
	}
	return &systemTransferReceiptStore{scopes: scopes, authorizer: authorizer,
		persistence: persistence}, nil
}

func (store *systemTransferReceiptStore) ExistingReceipts(ctx context.Context,
	attempt systemtransfer.Attempt, indexes []int64,
) (map[int64]systemtransfer.RowReceipt, error) {
	var result map[int64]systemtransfer.RowReceipt
	err := store.inFresh(ctx, attempt, func(scoped context.Context,
		scope platformtx.TenantScope, evidence publicresource.ActorEvidence) error {
		var err error
		result, err = store.persistence.LoadReceipts(
			scoped, scope, evidence, attempt, indexes)
		return err
	})
	return result, err
}

func (store *systemTransferReceiptStore) AppendReceipts(ctx context.Context,
	attempt systemtransfer.Attempt, receipts []systemtransfer.RowReceipt,
	sourceBytes int64,
) error {
	return store.inFresh(ctx, attempt, func(scoped context.Context,
		scope platformtx.TenantScope, evidence publicresource.ActorEvidence) error {
		return store.persistence.CommitReceipts(
			scoped, scope, evidence, attempt, receipts, sourceBytes)
	})
}

func (store *systemTransferReceiptStore) inFresh(ctx context.Context,
	attempt systemtransfer.Attempt,
	execute func(context.Context, platformtx.TenantScope,
		publicresource.ActorEvidence) error,
) error {
	if store == nil || attempt.Validate() != nil || execute == nil {
		return systemtransfer.ErrInvalidAttempt
	}
	return store.scopes.InFreshImportBatch(ctx, attempt,
		func(scoped context.Context, batch systemtransfer.BatchScope) error {
			evidence, err := store.authorizer.AuthorizeImportBatch(
				scoped, batch.IAMScope(), attempt)
			if err != nil {
				return err
			}
			return execute(scoped, batch.IAMScope(), evidence)
		})
}

var _ systemtransfer.ReceiptStore = (*systemTransferReceiptStore)(nil)
var _ systemTransferReceiptPersistence = (*systemtransferpostgres.WorkerStore)(nil)
