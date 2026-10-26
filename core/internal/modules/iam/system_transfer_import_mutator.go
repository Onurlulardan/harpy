package iam

import (
	"context"
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferResourceMutator struct {
	inner *identitySystemResourcePort
}

func newSystemTransferResourceMutator(pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	cursorKeys config.CursorKeyring,
	catalogs platformasync.WriteCatalogs,
) (*systemTransferResourceMutator, error) {
	if nilIdentityDependency(pool) || coordinator == nil {
		return nil, systemtransfer.ErrDependencies
	}
	identities, err := NewIdentityRuntime()
	if err != nil {
		return nil, err
	}
	queries, err := NewIAMRESTQueryRuntime()
	if err != nil {
		return nil, err
	}
	keys, err := identityQueryKeysFromConfig(cursorKeys)
	if err != nil {
		return nil, err
	}
	queryService, err := queries.newSystemResourceQuery(keys)
	if err != nil {
		return nil, err
	}
	evidence, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM import mutation evidence: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(evidence)
	if err != nil {
		return nil, fmt.Errorf("build IAM import mutation authorizer: %w", err)
	}
	changes, err := NewIdentityMutationSink(catalogs, time.Now)
	if err != nil {
		return nil, err
	}
	application := &identityHTTPApplication{
		pool: pool, coordinator: coordinator, identities: identities,
		queries: queryService, authorizer: authorizer, changes: changes,
	}
	return &systemTransferResourceMutator{
		inner: &identitySystemResourcePort{application: application},
	}, nil
}

func (mutator *systemTransferResourceMutator) PrepareCreate(ctx context.Context,
	input publicresource.CreateInput,
) (publicresource.PreparedCreate, error) {
	if mutator == nil || mutator.inner == nil {
		input.Destroy()
		return nil, publicresource.ErrUnavailable
	}
	resource, err := input.Resource()
	if err != nil || resource == publicresource.Users {
		input.Destroy()
		return nil, publicresource.ErrForbidden
	}
	return mutator.inner.PrepareCreate(ctx, input)
}

func (mutator *systemTransferResourceMutator) Create(ctx context.Context,
	scope platformtx.TenantScope,
	evidence publicresource.ActorEvidence,
	input publicresource.PreparedCreate,
) (publicresource.MutationResult, error) {
	if mutator == nil || mutator.inner == nil {
		input.Destroy()
		return publicresource.MutationResult{}, publicresource.ErrUnavailable
	}
	resource, err := input.Resource()
	if err != nil || resource == publicresource.Users {
		input.Destroy()
		return publicresource.MutationResult{}, publicresource.ErrForbidden
	}
	return mutator.inner.Create(ctx, scope, evidence, input)
}

func (mutator *systemTransferResourceMutator) Update(ctx context.Context,
	scope platformtx.TenantScope,
	evidence publicresource.ActorEvidence,
	input publicresource.UpdateInput,
) (publicresource.MutationResult, error) {
	if mutator == nil || mutator.inner == nil {
		return publicresource.MutationResult{}, publicresource.ErrUnavailable
	}
	return mutator.inner.Update(ctx, scope, evidence, input)
}

var _ systemtransfer.ResourceMutator = (*systemTransferResourceMutator)(nil)
