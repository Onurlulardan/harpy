package work

import (
	"context"

	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type principalResolver struct{ iam workprincipalapi.Directory }

type tenantScopeCarrier interface {
	PlatformTenantScope() platformtx.TenantScope
}

func newPrincipalResolver(runtime workprincipalapi.Directory) (*principalResolver, error) {
	if runtime == nil {
		return nil, ports.ErrUnavailable
	}
	return &principalResolver{iam: runtime}, nil
}

func (resolver *principalResolver) Resolve(ctx context.Context,
	local ports.Scope,
) (ports.PrincipalSnapshot, error) {
	scope, ok := platformScope(local)
	if resolver == nil || resolver.iam == nil || !ok {
		return ports.PrincipalSnapshot{}, ports.ErrUnavailable
	}
	evidence, err := resolver.iam.Evidence(ctx, scope, "USER", [16]byte(local.ActorID()))
	if err != nil || evidence.AuthorizationRevision < 1 {
		return ports.PrincipalSnapshot{}, ports.ErrUnavailable
	}
	if !evidence.TargetExists {
		return ports.PrincipalSnapshot{}, ports.ErrForbidden
	}
	principals := make([]ports.Principal, 0, len(evidence.Principals))
	for _, principal := range evidence.Principals {
		value := ports.Principal{Kind: ports.PrincipalKind(principal.Kind),
			ID: ports.ID(principal.ID)}
		if !value.Valid() || principal.Revision < 1 {
			return ports.PrincipalSnapshot{}, ports.ErrUnavailable
		}
		principals = append(principals, value)
	}
	if len(principals) == 0 {
		return ports.PrincipalSnapshot{}, ports.ErrForbidden
	}
	return ports.PrincipalSnapshot{Principals: principals,
		PermissionRevision: evidence.AuthorizationRevision}, nil
}

func (resolver *principalResolver) Exists(ctx context.Context, local ports.Scope,
	principal ports.Principal,
) (bool, int64, error) {
	scope, ok := platformScope(local)
	if resolver == nil || resolver.iam == nil || !ok || !principal.Valid() {
		return false, 0, ports.ErrUnavailable
	}
	evidence, err := resolver.iam.Evidence(ctx, scope,
		string(principal.Kind), [16]byte(principal.ID))
	if err != nil {
		return false, 0, ports.ErrUnavailable
	}
	return evidence.TargetExists, evidence.TargetRevision, nil
}

func platformScope(local ports.Scope) (platformtx.TenantScope, bool) {
	carrier, ok := local.(tenantScopeCarrier)
	if !ok {
		return platformtx.TenantScope{}, false
	}
	scope := carrier.PlatformTenantScope()
	return scope, scope.TenantID() != [16]byte{} && scope.ActorID() != [16]byte{}
}

var _ ports.PrincipalResolver = (*principalResolver)(nil)
