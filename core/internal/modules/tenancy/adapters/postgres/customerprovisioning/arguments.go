package customerprovisioning

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

func acceptArguments(plan tenantapi.OnboardingPlan, request provisioning.AcceptedRequest) []any {
	offering := plan.Offering()
	recovery := plan.RecoverySigning()
	authority := plan.Authority()
	return []any{
		uuid.UUID(plan.OperationID()), uuid.UUID(plan.EnvironmentID()), string(plan.Kind()),
		plan.IdempotencyKey(), plan.RequestSHA256(), plan.ManifestSHA256(),
		plan.StepPlanSHA256(), string(plan.DeploymentMode()), uuid.UUID(plan.TenantID()),
		plan.SchemaName(), string(plan.BindingPurpose()), uuid.UUID(plan.FirstUserID()),
		uuid.UUID(plan.RootOrganizationID()), uuid.UUID(plan.ProtectedAdminBindingID()),
		nullUUID(plan.InstanceOwnerBindingID()), uuid.UUID(offering.VersionID),
		offering.Revision, offering.SHA256, offering.EffectiveAt,
		nullInt64(offering.Source.Sequence, offering.Source.Present),
		nullString(offering.Source.SHA256, offering.Source.Present),
		nullInt64(recovery.Generation, recovery.Present),
		nullString(recovery.TrustBundleSHA256, recovery.Present),
		uuid.UUID(authority.RegistryBindingID), string(authority.ScopeKind),
		uuid.UUID(authority.ScopeID), authority.Generation,
		nullDigest(authority.PreviousHeadSHA256), nullDigest(authority.StateSHA256),
		nullDigest(authority.ProposalSHA256), plan.CreatedAt(), request.ControlTenantID,
		request.RequestedBySubjectID, request.Customer.Name, request.Customer.Slug,
		request.Customer.FirstAdmin.DisplayName, request.Customer.FirstAdmin.Email,
		request.JobID,
	}
}

func nullUUID(value tenantapi.UUID) any {
	if value == tenantapi.NilUUID {
		return nil
	}
	return uuid.UUID(value)
}
func nullInt64(value int64, present bool) any {
	if !present {
		return nil
	}
	return value
}
func nullString(value string, present bool) any {
	if !present {
		return nil
	}
	return value
}
func nullDigest(value string) any {
	if value == "" {
		return nil
	}
	return value
}
