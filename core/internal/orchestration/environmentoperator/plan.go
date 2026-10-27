package environmentoperator

import (
	"context"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) preparePlan(
	ctx context.Context,
	manifest installationv1.Manifest,
	createdAt time.Time,
	recovery tenantapi.RecoverySigningPin,
) (preparedPlan, error) {
	operationID := stableID(manifest, "operation")
	environmentID := uuid.UUID(manifest.EnvironmentID())
	tenantID := stableID(manifest, "primary-tenant")
	mode, purpose, err := onboardingMode(manifest.DeploymentMode())
	if err != nil {
		return preparedPlan{}, err
	}
	offering, err := coordinator.offering.Resolve(ctx, OfferingRequest{
		Manifest: manifest, OperationID: operationID, CreatedAt: createdAt,
	})
	if err != nil {
		return preparedPlan{}, classifyDependency(err)
	}
	schema, err := tenantapi.OnboardingSchemaName(tenantID)
	if err != nil {
		return preparedPlan{}, ErrInvalid
	}
	authorityRequest := AuthorityRequest{OperationID: operationID,
		EnvironmentID: environmentID, TenantID: tenantID,
		FirstUserID:             stableID(manifest, "first-user"),
		RootOrganizationID:      stableID(manifest, "root-organization"),
		ProtectedAdminBindingID: stableID(manifest, "protected-admin-binding"),
		InstanceOwnerBindingID:  stableID(manifest, "instance-owner-binding"),
		RegistryBindingID:       stableID(manifest, "authority-registry-binding"),
		DeploymentMode:          mode, RecoverySigning: recovery,
		IssuedAt: createdAt.Truncate(time.Second),
	}
	authority, err := coordinator.prepareAuthority(ctx, authorityRequest)
	if err != nil {
		return preparedPlan{}, err
	}
	plan, err := tenantapi.NewOnboardingPlan(tenantapi.OnboardingPlanInput{
		OperationID: operationID, EnvironmentID: environmentID, TenantID: tenantID,
		FirstUserID:             authorityRequest.FirstUserID,
		RootOrganizationID:      authorityRequest.RootOrganizationID,
		ProtectedAdminBindingID: authorityRequest.ProtectedAdminBindingID,
		InstanceOwnerBindingID:  authorityRequest.InstanceOwnerBindingID,
		Kind:                    tenantapi.OnboardingEnvironmentInitialize, DeploymentMode: mode,
		BindingPurpose: purpose, SchemaName: schema,
		IdempotencyKey: "environment_initialize:" + manifest.EnvironmentIDString(),
		RequestSHA256:  manifest.RequestSHA256(), ManifestSHA256: manifest.SHA256(),
		Offering: offering, RecoverySigning: recovery, Authority: authority,
		CreatedAt: createdAt,
	})
	if err != nil {
		return preparedPlan{}, ErrDependencyMismatch
	}
	return preparedPlan{plan: plan, authorityRequest: authorityRequest}, nil
}

func onboardingMode(mode installationv1.DeploymentMode) (
	tenantapi.DeploymentMode, tenantapi.TenantBindingPurpose, error,
) {
	switch mode {
	case installationv1.DeploymentOnPremise:
		return tenantapi.DeploymentModeOnPremise,
			tenantapi.TenantBindingOnPremisePrimary, nil
	case installationv1.DeploymentSaaS:
		return tenantapi.DeploymentModeSaaS, tenantapi.TenantBindingSaaSControl, nil
	default:
		return "", "", ErrInvalid
	}
}
