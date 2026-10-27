package audit

import (
	"context"
	"fmt"
	"reflect"

	audithttp "github.com/Onurlulardan/harpy/core/internal/modules/audit/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type HTTPGenerationResolver interface {
	ResolveTenantAuditGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type HTTPScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type HTTPRuntime struct {
	resolver    HTTPGenerationResolver
	coordinator HTTPScopeCoordinator
	application *application.Application
}

func NewHTTPRuntime(
	resolver HTTPGenerationResolver, coordinator HTTPScopeCoordinator,
	core *application.Application,
) (*HTTPRuntime, error) {
	if nilHTTPPort(resolver) || nilHTTPPort(coordinator) || core == nil {
		return nil, fmt.Errorf("audit HTTP runtime dependencies are invalid")
	}
	return &HTTPRuntime{resolver: resolver, coordinator: coordinator,
		application: core}, nil
}

func (runtime *HTTPRuntime) Ready() bool {
	return runtime != nil && !nilHTTPPort(runtime.resolver) &&
		!nilHTTPPort(runtime.coordinator) && runtime.application != nil
}

func (runtime *HTTPRuntime) List(
	ctx context.Context, access audithttp.Access, filter audithttp.Filter, operation string,
) (result audithttp.Page, err error) {
	err = runtime.snapshot(ctx, access, operation, func(authorized context.Context,
		scope platformtx.TenantScope,
	) error {
		result, err = runtime.application.List(authorized, scope, filter, operation)
		return err
	})
	return result, err
}

func (runtime *HTTPRuntime) Stats(
	ctx context.Context, access audithttp.Access,
) (result audithttp.Stats, err error) {
	err = runtime.snapshot(ctx, access, application.OperationStats,
		func(authorized context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.Stats(authorized, scope)
			return err
		})
	return result, err
}

func (runtime *HTTPRuntime) Get(
	ctx context.Context, access audithttp.Access, id uuid.UUID,
) (result audithttp.Entry, err error) {
	err = runtime.snapshot(ctx, access, application.OperationGet,
		func(authorized context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.Get(authorized, scope, id)
			return err
		})
	return result, err
}

func (runtime *HTTPRuntime) snapshot(
	ctx context.Context, access audithttp.Access, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if runtime == nil || ctx == nil || ctx.Err() != nil || action == nil ||
		!access.ValidFor(application.PermissionRead, operation) {
		return application.ErrForbidden
	}
	group, generation, err := runtime.resolver.ResolveTenantAuditGenerations(ctx, access.TenantID())
	if err != nil || group < 1 || generation < 1 {
		return application.ErrUnavailable
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return application.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		operation, map[string]int64{"audit": generation})
	if err != nil {
		return application.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery},
		func(scoped context.Context, scope platformtx.TenantScope) error {
			evidence, evidenceErr := application.NewEvidence(access.PrincipalID(), access.TenantID(),
				access.TokenVersion(), access.AuthorizationVersion(), access.AuthorizationClock(),
				string(application.PermissionRead), operation)
			if evidenceErr != nil {
				return application.ErrUnavailable
			}
			authorized, evidenceErr := application.WithEvidence(scoped, evidence)
			if evidenceErr != nil {
				return application.ErrUnavailable
			}
			return action(authorized, scope)
		})
}

func nilHTTPPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

var _ audithttp.Application = (*HTTPRuntime)(nil)
