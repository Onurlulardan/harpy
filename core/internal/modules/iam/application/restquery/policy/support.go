package policy

import (
	"context"
	"reflect"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	iamauthorization "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type AuthorizationPort interface {
	Authorize(
		context.Context,
		transaction.TenantScope,
		iamauthorization.Request,
	) (iamauthorization.Decision, error)
}

type CatalogPort interface {
	Catalog(
		context.Context,
		transaction.TenantScope,
		profile.ResourceKind,
	) (catalog.Published, error)
}

type VersionPort interface {
	Versions(
		context.Context,
		transaction.TenantScope,
		uuid.UUID,
	) (ActorVersions, error)
}

type Service struct {
	authorizer AuthorizationPort
	catalogs   CatalogPort
	versions   VersionPort
}

func NewService(
	authorizer AuthorizationPort,
	catalogs CatalogPort,
	versions VersionPort,
) (*Service, error) {
	if interfaceNil(authorizer) || interfaceNil(catalogs) || interfaceNil(versions) {
		return nil, catalog.ErrContract
	}
	return &Service{authorizer: authorizer, catalogs: catalogs, versions: versions}, nil
}

func interfaceNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

var _ restauthorization.PolicyPort = (*Service)(nil)
