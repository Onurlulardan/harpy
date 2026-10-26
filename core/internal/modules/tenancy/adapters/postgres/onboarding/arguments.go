package onboarding

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func onboardingAcceptArguments(plan tenantapi.OnboardingPlan) []any {
	offering := plan.Offering()
	recovery := plan.RecoverySigning()
	authority := plan.Authority()
	return []any{
		uuid.UUID(plan.OperationID()), uuid.UUID(plan.EnvironmentID()),
		string(plan.Kind()), plan.IdempotencyKey(),
		plan.RequestSHA256(), plan.ManifestSHA256(), plan.StepPlanSHA256(),
		string(plan.DeploymentMode()), uuid.UUID(plan.TenantID()), plan.SchemaName(),
		string(plan.BindingPurpose()), uuid.UUID(plan.FirstUserID()),
		uuid.UUID(plan.RootOrganizationID()), uuid.UUID(plan.ProtectedAdminBindingID()),
		optionalUUID(plan.InstanceOwnerBindingID()),
		uuid.UUID(offering.VersionID), offering.Revision, offering.SHA256, offering.EffectiveAt,
		optionalInt64(offering.Source.Sequence, offering.Source.Present),
		optionalString(offering.Source.SHA256, offering.Source.Present),
		optionalInt64(recovery.Generation, recovery.Present),
		optionalString(recovery.TrustBundleSHA256, recovery.Present),
		uuid.UUID(authority.RegistryBindingID), string(authority.ScopeKind),
		uuid.UUID(authority.ScopeID),
		authority.Generation, optionalDigest(authority.PreviousHeadSHA256),
		optionalDigest(authority.StateSHA256), optionalDigest(authority.ProposalSHA256),
		plan.CreatedAt(),
	}
}

func optionalUUID(value tenantapi.UUID) any {
	if value == tenantapi.NilUUID {
		return nil
	}
	return uuid.UUID(value)
}

func optionalInt64(value int64, present bool) any {
	if !present {
		return nil
	}
	return value
}

func optionalString(value string, present bool) any {
	if !present {
		return nil
	}
	return value
}

func optionalDigest(value string) any {
	if value == "" {
		return nil
	}
	return value
}
