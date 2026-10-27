package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	publicapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *identitySystemResourcePort) Catalog(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	resource publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	evidence, err := newIdentityPublicEvidence(actor)
	if err != nil || identitySystemResourceScope(
		ctx, scope, evidence, platformtx.SnapshotQuery, "restquery.execute",
	) != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	return port.authorizedCatalog(ctx, scope, evidence, resource)
}

func (port *identitySystemResourcePort) CatalogForLayoutPublish(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	resource publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	evidence, err := newIdentityPublicEvidence(actor)
	if err != nil ||
		identitySystemResourceLayoutPublishScope(ctx, scope, evidence) != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	return port.authorizedCatalog(ctx, scope, evidence, resource)
}

func (port *identitySystemResourcePort) authorizedCatalog(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	resource publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	authority, err := newIdentitySystemResourceAuthorization(
		port.application, ctx, scope, evidence, resource,
	)
	if err != nil {
		return publicresource.FieldCatalog{}, err
	}
	if _, err = authority.require("read"); err != nil {
		return publicresource.FieldCatalog{}, err
	}
	return port.resolveCatalog(ctx, scope, resource, authority)
}

func (port *identitySystemResourcePort) resolveCatalog(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource publicresource.Resource,
	authority *identitySystemResourceAuthorization,
) (publicresource.FieldCatalog, error) {
	result, _, err := port.resolveCatalogSnapshot(ctx, scope, resource, authority)
	return result, err
}

func (port *identitySystemResourcePort) resolveCatalogSnapshot(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource publicresource.Resource,
	authority *identitySystemResourceAuthorization,
) (publicresource.FieldCatalog, catalog.Published, error) {
	confidential, err := authority.allowed("read.confidential")
	if err != nil {
		return publicresource.FieldCatalog{}, catalog.Published{}, err
	}
	restricted, err := authority.allowed("read.restricted")
	if err != nil {
		return publicresource.FieldCatalog{}, catalog.Published{}, err
	}
	visibility, err := publicapp.NewCatalogVisibility(true, confidential, restricted)
	if err != nil {
		return publicresource.FieldCatalog{}, catalog.Published{}, publicresource.ErrContract
	}
	published, err := port.application.queries.ResolveCatalog(
		ctx, scope, string(resource),
	)
	if err != nil {
		return publicresource.FieldCatalog{}, catalog.Published{}, identitySystemResourceError(err)
	}
	fieldCatalog, err := publicapp.BuildFieldCatalog(published, visibility)
	if err != nil {
		return publicresource.FieldCatalog{}, catalog.Published{}, publicresource.ErrUnavailable
	}
	return fieldCatalog, published, nil
}
