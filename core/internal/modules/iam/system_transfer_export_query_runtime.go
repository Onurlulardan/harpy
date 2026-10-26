package iam

import (
	"context"

	policypostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/restquery/policy"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	querycatalog "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/policy"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type systemTransferPinnedCatalogPort struct {
	kind      profileapp.ResourceKind
	published querycatalog.Published
}

func (port systemTransferPinnedCatalogPort) Catalog(
	ctx context.Context,
	scope platformtx.TenantScope,
	kind profileapp.ResourceKind,
) (querycatalog.Published, error) {
	if !systemTransferExportSnapshotScopeUsable(ctx, scope) || kind != port.kind ||
		port.published.Validate() != nil {
		return querycatalog.Published{}, querycatalog.ErrContract
	}
	return port.published, nil
}

type systemTransferPinnedVersionPort struct {
	actorID                     uuid.UUID
	tokenVersion, authorization int64
}

func (port systemTransferPinnedVersionPort) Versions(
	ctx context.Context,
	scope platformtx.TenantScope,
	actorID uuid.UUID,
) (policyapp.ActorVersions, error) {
	if !systemTransferExportSnapshotScopeUsable(ctx, scope) || actorID != port.actorID ||
		actorID != scope.ActorID() || port.tokenVersion < 1 || port.authorization < 1 {
		return policyapp.ActorVersions{}, querycatalog.ErrContract
	}
	return policyapp.NewActorVersions(actorID, port.tokenVersion, port.authorization)
}

func newPinnedSystemResourceQuery(
	keys queryapp.CursorKeys,
	plan systemTransferExportPlan,
	attempt systemtransfer.ExportAttempt,
) (*queryapp.Service, error) {
	kind, err := systemTransferExportResourceKind(attempt.Resource)
	if err != nil || plan.published.Validate() != nil ||
		!systemTransferExportPinMatches(plan.published, attempt) {
		return nil, systemtransfer.ErrExportCatalogPinStale
	}
	catalogs := systemTransferPinnedCatalogPort{kind: kind, published: plan.published}
	versions := systemTransferPinnedVersionPort{actorID: attempt.ActorID,
		tokenVersion: attempt.TokenVersion, authorization: attempt.AuthorizationVersion}
	policy, err := policypostgres.NewPolicyServiceForSnapshot(catalogs, versions)
	if err != nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	users, roles, organizations, err := buildIAMRESTQueryCatalogs()
	if err != nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	execution, err := buildIAMRESTQueryExecution()
	if err != nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	service, err := queryapp.NewService(
		users, roles, organizations, policy, execution, keys,
	)
	if err != nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	return service, nil
}

func systemTransferExportSnapshotScopeUsable(
	ctx context.Context,
	scope platformtx.TenantScope,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil &&
		scope.OperationID() == "restquery.execute"
}

var _ policyapp.CatalogPort = systemTransferPinnedCatalogPort{}
var _ policyapp.VersionPort = systemTransferPinnedVersionPort{}
