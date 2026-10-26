package organization

import (
	"context"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Port interface {
	List(context.Context, transaction.TenantScope) (catalog.Published, error)
}

type Service struct{ port Port }

func NewService(port Port) (*Service, error) {
	if portIsNil(port) {
		return nil, catalog.ErrContract
	}
	return &Service{port: port}, nil
}

func scopeUsable(ctx context.Context, scope transaction.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		catalogReadScopeActive(scope)
}

func catalogReadScopeActive(scope transaction.TenantScope) bool {
	return transaction.RequireActive(scope, transaction.SnapshotQuery) == nil ||
		transaction.RequireActive(scope, transaction.Command) == nil
}

func portIsNil(port Port) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
