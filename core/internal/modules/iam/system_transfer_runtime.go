package iam

import (
	"context"
	"fmt"
	"time"

	systemtransferpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfer"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type SystemTransferScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type SystemTransferRuntimeDependencies struct {
	Generations IdentityHTTPPlatform
	Scopes      SystemTransferScopeCoordinator
	Queries     *queryapp.Service
	Resources   publicresource.Port
	Controls    systemtransfer.ControlStoreBinder
	Catalogs    platformasync.WriteCatalogs
	Artifacts   SystemTransferArtifactRuntime
	Clock       func() time.Time
	NewID       func() (uuid.UUID, error)
}

type SystemTransferRuntime struct {
	generations IdentityHTTPPlatform
	scopes      SystemTransferScopeCoordinator
	queries     *queryapp.Service
	resources   publicresource.Port
	controls    systemtransfer.ControlStoreBinder
	catalogs    platformasync.WriteCatalogs
	artifacts   SystemTransferArtifactRuntime
	clock       func() time.Time
	newID       func() (uuid.UUID, error)
}

func NewSystemTransferPlatformRuntime(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	identity IdentityHTTPApplication,
	queries *IAMRESTQueryRuntime,
	cursorKeys config.CursorKeyring,
	catalogs platformasync.WriteCatalogs,
	artifacts SystemTransferArtifactRuntime,
) (SystemTransferApplication, error) {
	if nilIdentityDependency(identity) || queries == nil || !queries.Valid() {
		return nil, fmt.Errorf("IAM transfer platform dependencies are incomplete")
	}
	keys, err := identityQueryKeysFromConfig(cursorKeys)
	if err != nil {
		return nil, err
	}
	queryService, err := queries.newSystemResourceQuery(keys)
	if err != nil {
		return nil, err
	}
	return NewSystemTransferRuntime(SystemTransferRuntimeDependencies{
		Generations: pool, Scopes: coordinator, Queries: queryService,
		Resources: identity.SystemResources(), Controls: systemtransferpostgres.NewBinder(),
		Catalogs: catalogs, Artifacts: artifacts, Clock: time.Now, NewID: uuid.NewV7,
	})
}

func NewSystemTransferRuntime(
	dependencies SystemTransferRuntimeDependencies,
) (*SystemTransferRuntime, error) {
	if nilIdentityDependency(dependencies.Generations) ||
		nilIdentityDependency(dependencies.Scopes) || dependencies.Queries == nil ||
		nilIdentityDependency(dependencies.Resources) ||
		nilIdentityDependency(dependencies.Controls) ||
		nilIdentityDependency(dependencies.Artifacts) || dependencies.Clock == nil ||
		dependencies.NewID == nil {
		return nil, ErrSystemTransferUnavailable
	}
	if jobs, _, err := dependencies.Catalogs.Snapshot(); err != nil {
		return nil, ErrSystemTransferUnavailable
	} else if export, err := jobs.Resolve(SystemExportJobKind, 1); err != nil ||
		ValidateSystemExportJobPolicy(export) != nil {
		return nil, ErrSystemTransferUnavailable
	} else if imp, err := jobs.Resolve(SystemImportJobKind, 1); err != nil ||
		ValidateSystemImportJobPolicy(imp) != nil {
		return nil, ErrSystemTransferUnavailable
	}
	return &SystemTransferRuntime{
		generations: dependencies.Generations, scopes: dependencies.Scopes,
		queries: dependencies.Queries, resources: dependencies.Resources,
		controls: dependencies.Controls, catalogs: dependencies.Catalogs,
		artifacts: dependencies.Artifacts, clock: dependencies.Clock,
		newID: dependencies.NewID,
	}, nil
}

var _ SystemTransferApplication = (*SystemTransferRuntime)(nil)
