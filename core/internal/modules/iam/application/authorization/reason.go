package authorization

type Reason string

const (
	ReasonInvalidRequest                   Reason = "invalid_request"
	ReasonEvidenceUnavailable              Reason = "evidence_unavailable"
	ReasonActorUnavailable                 Reason = "actor_unavailable"
	ReasonTokenStale                       Reason = "token_stale"
	ReasonAuthorizationStale               Reason = "authorization_stale"
	ReasonResourceUnavailable              Reason = "resource_unavailable"
	ReasonActionUnavailable                Reason = "action_unavailable"
	ReasonNoEffectivePermission            Reason = "no_effective_permission"
	ReasonDirectUserExact                  Reason = "direct_user_exact"
	ReasonRoleExact                        Reason = "role_exact"
	ReasonOrganizationExact                Reason = "organization_exact"
	ReasonProtectedResourceActions Reason = "protected_resource_all_actions"
	ReasonProtectedTenantAll       Reason = "protected_tenant_all"
)

func (reason Reason) allowed() bool {
	switch reason {
	case ReasonDirectUserExact, ReasonRoleExact, ReasonOrganizationExact,
		ReasonProtectedResourceActions, ReasonProtectedTenantAll:
		return true
	default:
		return false
	}
}

func (reason Reason) denied() bool {
	switch reason {
	case ReasonInvalidRequest, ReasonEvidenceUnavailable, ReasonActorUnavailable,
		ReasonTokenStale, ReasonAuthorizationStale, ReasonResourceUnavailable,
		ReasonActionUnavailable, ReasonNoEffectivePermission:
		return true
	default:
		return false
	}
}
