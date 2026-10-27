package iam

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferExportSnapshots struct {
	clock      systemTransferArtifactClock
	scopes     *systemTransferExportScopes
	authorizer *systemTransferExportAuthorizer
	keys       queryapp.CursorKeys
}

func newSystemTransferExportSnapshots(
	scopes *systemTransferExportScopes,
	authorizer *systemTransferExportAuthorizer,
	keys queryapp.CursorKeys,
) (*systemTransferExportSnapshots, error) {
	if scopes == nil || authorizer == nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	return &systemTransferExportSnapshots{
		clock: systemTransferArtifactClock{}, scopes: scopes,
		authorizer: authorizer, keys: keys,
	}, nil
}

func (provider *systemTransferExportSnapshots) WithPinnedSnapshot(
	ctx context.Context,
	attempt systemtransfer.ExportAttempt,
	execute func(context.Context, systemtransfer.ExportSnapshot) error,
) error {
	if provider == nil || provider.scopes == nil || provider.authorizer == nil ||
		ctx == nil || ctx.Err() != nil || attempt.Validate() != nil || execute == nil {
		return systemtransfer.ErrExportWorkerInvalid
	}
	invoked := false
	err := provider.scopes.withSnapshot(ctx, attempt,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if invoked || !systemTransferExportSnapshotScopeUsable(scoped, scope) ||
				scope.TenantID() != attempt.TenantID || scope.ActorID() != attempt.ActorID {
				return systemtransfer.ErrExportStateConflict
			}
			invoked = true
			if _, err := provider.authorizer.authorizeSnapshot(scoped, scope, attempt); err != nil {
				return err
			}
			plan, err := loadSystemTransferExportPlan(scoped, scope, attempt)
			if err != nil {
				return err
			}
			service, err := newPinnedSystemResourceQuery(provider.keys, plan, attempt)
			if err != nil {
				return err
			}
			rows, err := newSystemTransferExportRows(scoped, scope, service, plan)
			if err != nil {
				return err
			}
			return execute(scoped, systemtransfer.ExportSnapshot{
				CatalogVersionID: attempt.CatalogVersionID,
				CatalogSHA256:    attempt.CatalogSHA256, QuerySHA256: attempt.QuerySHA256,
				StartedAt: provider.clock.Now(), Columns: append([]string(nil), plan.columns...),
				Rows: rows,
			})
		})
	if err != nil {
		return err
	}
	if !invoked {
		return systemtransfer.ErrExportStateConflict
	}
	return nil
}

var _ systemtransfer.ExportSnapshotProvider = (*systemTransferExportSnapshots)(nil)
