package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/authority"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) snapshot(ctx context.Context, access cataloghttp.Access,
	permission cataloghttp.Permission, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	_, catalogGeneration, err := runtime.resolveCatalog(ctx, access, permission, operation)
	if err != nil {
		return err
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"catalog": catalogGeneration})
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	return runtime.scopes.InTenant(ctx, platformtx.ScopeRequest{Actor: actor,
		Operation: verified, Mode: platformtx.SnapshotQuery}, runtime.authorized(access, action))
}

func (runtime *HTTPRuntime) menuSnapshot(ctx context.Context, access cataloghttp.Access,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	generations, err := runtime.resolveMenuGenerations(
		ctx, access, cataloghttp.PermissionRead, "catalog.menu.get",
	)
	if err != nil {
		return err
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		"catalog.menu.get", generations,
	)
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	return runtime.scopes.InTenant(ctx, platformtx.ScopeRequest{Actor: actor,
		Operation: verified, Mode: platformtx.SnapshotQuery}, runtime.authorized(access, action))
}

func (runtime *HTTPRuntime) catalogCommand(ctx context.Context, access cataloghttp.Access,
	permission cataloghttp.Permission, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	group, generation, err := runtime.resolveCatalog(ctx, access, permission, operation)
	if err != nil {
		return err
	}
	request, err := platformtx.AdmitCommandScope(access.PrincipalID(), access.TenantID(),
		operation, "catalog", group, map[string]int64{"catalog": generation})
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	return runtime.scopes.InTenant(ctx, request, runtime.authorized(access, action))
}

func (runtime *HTTPRuntime) lifecycleCommand(ctx context.Context, access cataloghttp.Access,
	permission cataloghttp.Permission, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	groups, domains, err := runtime.resolveLifecycle(ctx, access, permission, operation)
	if err != nil {
		return err
	}
	request, err := platformtx.AdmitMultiGroupCommandScope(access.PrincipalID(), access.TenantID(),
		operation, groups, domains, []string{"analytics", "catalog", "iam", "workflow"})
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	return runtime.scopes.InTenant(ctx, request, runtime.authorized(access, action))
}

func (runtime *HTTPRuntime) authorized(access cataloghttp.Access,
	action func(context.Context, platformtx.TenantScope) error,
) func(context.Context, platformtx.TenantScope) error {
	return func(ctx context.Context, scope platformtx.TenantScope) error {
		evidence, err := catalogAuthorityEvidence(access)
		if err != nil {
			return cataloghttp.ErrUnavailable
		}
		authorized, err := authority.WithEvidence(ctx, evidence)
		if err != nil {
			return cataloghttp.ErrUnavailable
		}
		tokenVersion, authorizationVersion, authorizationClock,
			permissionID, authorizationSHA256, purposeSHA256,
			hasReceipt := evidence.AuthorizationReceipt()
		if hasReceipt && attachLifecycleAuthorizationReceipt(
			access.Permission(), scope.OperationID(),
		) {
			if !evidence.Authorizes(scope, string(access.Permission())) {
				return cataloghttp.ErrForbidden
			}
			receipt, receiptErr := lifecycleapp.NewAuthorizationReceipt(
				scope, string(access.Permission()), tokenVersion,
				authorizationVersion, authorizationClock, permissionID,
				authorizationSHA256, purposeSHA256,
			)
			if receiptErr != nil {
				return cataloghttp.ErrUnavailable
			}
			authorized, err = lifecycleapp.WithAuthorizationReceipt(
				authorized, receipt,
			)
			if err != nil {
				return cataloghttp.ErrUnavailable
			}
		}
		return action(authorized, scope)
	}
}

func attachLifecycleAuthorizationReceipt(
	permission cataloghttp.Permission,
	operation string,
) bool {
	return permission == cataloghttp.PermissionPurge &&
		operation == "catalog.installation.purge"
}

func catalogAuthorityEvidence(access cataloghttp.Access) (authority.Evidence, error) {
	if access.AuthorizationPermissionID() == uuid.Nil {
		return authority.NewEvidence(
			access.PrincipalID(), access.TenantID(), access.TokenVersion(),
			access.AuthorizationVersion(), access.AuthorizationClock(),
			string(access.Permission()), access.Operation(), access.StepUpAt(),
		)
	}
	return authority.NewEvidenceWithAuthorizationReceipt(
		access.PrincipalID(), access.TenantID(),
		access.TokenVersion(), access.AuthorizationVersion(), access.AuthorizationClock(),
		string(access.Permission()), access.Operation(), access.StepUpAt(),
		access.AuthorizationPermissionID(), access.AuthorizationEvidenceSHA256(),
		access.StepUpPurposeSHA256(),
	)
}
