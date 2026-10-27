package customerprovisioning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type buildPlanInput struct {
	EnvironmentID  uuid.UUID
	IdempotencyKey string
	RequestSHA256  string
	Offering       tenantapi.BaseOfferingPin
	CreatedAt      time.Time
}

func (coordinator *Coordinator) buildPlan(
	ctx context.Context,
	input buildPlanInput,
) (tenantapi.OnboardingPlan, error) {
	operation := operationID(input.EnvironmentID, input.IdempotencyKey)
	tenantID := stableID(operation, "tenant")
	firstUserID := stableID(operation, "first-user")
	protectedAdminID := stableID(operation, "protected-admin-binding")
	authority, err := coordinator.dependencies.Authority.PrepareInitialTenant(
		ctx, AuthorityRequest{OperationID: operation, EnvironmentID: input.EnvironmentID,
			TenantID: tenantID, FirstUserID: firstUserID,
			ProtectedAdminBindingID: protectedAdminID,
			IssuedAt:                input.CreatedAt.Truncate(time.Second)},
	)
	if err != nil || !validAuthority(authority, tenantID) {
		return tenantapi.OnboardingPlan{}, ErrUnavailable
	}
	schemaName, err := tenantapi.OnboardingSchemaName(tenantapi.UUID(tenantID))
	if err != nil {
		return tenantapi.OnboardingPlan{}, ErrInvalid
	}
	manifestSHA, err := customerManifestSHA(input, operation, tenantID,
		firstUserID, protectedAdminID, authority)
	if err != nil {
		return tenantapi.OnboardingPlan{}, ErrUnavailable
	}
	plan, err := tenantapi.NewOnboardingPlan(tenantapi.OnboardingPlanInput{
		OperationID: tenantapi.UUID(operation), EnvironmentID: tenantapi.UUID(input.EnvironmentID),
		TenantID: tenantapi.UUID(tenantID), FirstUserID: tenantapi.UUID(firstUserID),
		RootOrganizationID:      tenantapi.UUID(stableID(operation, "root-organization")),
		ProtectedAdminBindingID: tenantapi.UUID(protectedAdminID),
		InstanceOwnerBindingID:  tenantapi.NilUUID,
		Kind:                    tenantapi.OnboardingSaaSCustomerProvision,
		DeploymentMode:          tenantapi.DeploymentModeSaaS,
		BindingPurpose:          tenantapi.TenantBindingSaaSCustomer, SchemaName: schemaName,
		IdempotencyKey: input.IdempotencyKey, RequestSHA256: input.RequestSHA256,
		ManifestSHA256: manifestSHA, Offering: input.Offering,
		Authority: authority, CreatedAt: input.CreatedAt,
	})
	if err != nil {
		return tenantapi.OnboardingPlan{}, ErrInvalid
	}
	return plan, nil
}

func validAuthority(value tenantapi.ProtectedAuthorityPlan, tenantID uuid.UUID) bool {
	return value.RegistryBindingID != tenantapi.NilUUID &&
		value.ScopeKind == tenantapi.AuthorityScopeTenant &&
		value.ScopeID == tenantapi.UUID(tenantID) && value.Generation == 1 &&
		value.PreviousHeadSHA256 == "" && len(value.StateSHA256) == 64 &&
		len(value.ProposalSHA256) == 64
}

func customerManifestSHA(input buildPlanInput, operation, tenantID, firstUserID,
	protectedAdminID uuid.UUID, authority tenantapi.ProtectedAuthorityPlan,
) (string, error) {
	document := struct {
		Schema, OperationID, EnvironmentID, TenantID, FirstUserID string
		ProtectedAdminID, RequestSHA256, OfferingSHA256           string
		AuthorityStateSHA256, AuthorityProposalSHA256             string
	}{"harpy-saas-customer-provision.v1", operation.String(),
		input.EnvironmentID.String(), tenantID.String(), firstUserID.String(),
		protectedAdminID.String(), input.RequestSHA256, input.Offering.SHA256,
		authority.StateSHA256, authority.ProposalSHA256}
	wire, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(wire)
	return hex.EncodeToString(digest[:]), nil
}
