package iam

import (
	"context"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferPreparedMapping struct {
	document       []byte
	digest         [32]byte
	manifestDigest [32]byte
	fieldCount     int
	catalog        systemTransferCatalogPin
}

func (runtime *SystemTransferRuntime) prepareSystemImportMapping(
	ctx context.Context,
	scope platformtx.TenantScope,
	access SystemTransferAccess,
	resource publicresource.Resource,
	document []byte,
) (systemTransferPreparedMapping, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 1<<20)
	if err != nil {
		return systemTransferPreparedMapping{}, ErrSystemTransferInvalid
	}
	actor, err := systemTransferActor(access)
	if err != nil {
		return systemTransferPreparedMapping{}, err
	}
	catalog, err := runtime.resources.Catalog(ctx, scope, actor, resource)
	if err != nil {
		return systemTransferPreparedMapping{}, systemTransferError(err)
	}
	policy, err := systemtransfer.NewMappingPolicy(catalog)
	if err != nil {
		return systemTransferPreparedMapping{}, ErrSystemTransferInvalid
	}
	program, err := policy.Compile(canonical)
	if err != nil || len(program.Targets()) < 1 || len(program.Targets()) > 500 {
		return systemTransferPreparedMapping{}, ErrSystemTransferInvalid
	}
	manifest, err := systemtransfer.ManifestFor(resource)
	if err != nil {
		return systemTransferPreparedMapping{}, ErrSystemTransferInvalid
	}
	pin, err := systemTransferPublicPin(catalog)
	if err != nil {
		return systemTransferPreparedMapping{}, err
	}
	return systemTransferPreparedMapping{
		document: canonical, digest: systemTransferSHA256(canonical),
		manifestDigest: manifest.Digest, fieldCount: len(program.Targets()), catalog: pin,
	}, nil
}

func (runtime *SystemTransferRuntime) currentSystemTransferPin(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource SystemTransferResource,
) (systemTransferCatalogPin, error) {
	published, err := runtime.queries.ResolveCatalog(ctx, scope, string(resource))
	if err != nil {
		return systemTransferCatalogPin{}, systemTransferError(err)
	}
	return systemTransferPublishedPin(published)
}
