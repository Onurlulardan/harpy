package platformiam

import (
	"context"
	"reflect"
	"time"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }

type Service struct {
	store platformports.Store[platformtx.TenantScope]
	clock Clock
}

func NewService(
	store platformports.Store[platformtx.TenantScope], clock Clock,
) (*Service, error) {
	if nilPort(store) || nilPort(clock) {
		return nil, platformdomain.ErrStorage
	}
	return &Service{store: store, clock: clock}, nil
}

func commandScope(
	ctx context.Context, scope platformtx.TenantScope, actor platformdomain.ID,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		scope.ActorID() == uuid.UUID(actor)
}

func queryScope(
	ctx context.Context, scope platformtx.TenantScope, actor platformdomain.ID,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil &&
		scope.ActorID() == uuid.UUID(actor)
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}
