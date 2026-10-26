package customerworker

import (
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
	"github.com/google/uuid"
)

func scanClaim(row Row, leaseTokenSHA256 string) (workerapp.Claim, error) {
	var operationID, environmentID, tenantID uuid.UUID
	var firstUserID, rootOrganizationID, protectedAdminID uuid.UUID
	var offeringVersionID, registryBindingID, authorityScopeID uuid.UUID
	var instanceOwner *uuid.UUID
	var sourceSequence, recoveryGeneration *int64
	var sourceSHA, recoverySHA, previousHead *string
	var operationKind, idempotencyKey, requestSHA, manifestSHA, stepPlanSHA string
	var deploymentMode, schemaName, bindingPurpose string
	var offeringRevision, authorityGeneration, revision int64
	var offeringSHA, authorityKind, authorityState, authorityProposal string
	var customerName, tenantSlug, adminDisplayName, adminEmail, state string
	var offeringEffectiveAt, createdAt, leaseUntil time.Time
	err := row.Scan(&operationID, &environmentID, &operationKind,
		&idempotencyKey, &requestSHA, &manifestSHA, &stepPlanSHA,
		&deploymentMode, &tenantID, &schemaName, &bindingPurpose,
		&firstUserID, &rootOrganizationID, &protectedAdminID, &instanceOwner,
		&offeringVersionID, &offeringRevision, &offeringSHA, &offeringEffectiveAt,
		&sourceSequence, &sourceSHA, &recoveryGeneration, &recoverySHA,
		&registryBindingID, &authorityKind, &authorityScopeID,
		&authorityGeneration, &previousHead, &authorityState, &authorityProposal,
		&createdAt, &customerName, &tenantSlug, &adminDisplayName, &adminEmail,
		&state, &revision, &leaseUntil)
	if err != nil {
		return workerapp.Claim{}, err
	}
	if instanceOwner != nil || sourceSequence != nil || sourceSHA != nil ||
		recoveryGeneration != nil || recoverySHA != nil || previousHead != nil {
		return workerapp.Claim{}, workerapp.ErrInvalid
	}
	plan, err := tenantapi.NewOnboardingPlan(tenantapi.OnboardingPlanInput{
		OperationID: tenantapi.UUID(operationID), EnvironmentID: tenantapi.UUID(environmentID),
		TenantID: tenantapi.UUID(tenantID), FirstUserID: tenantapi.UUID(firstUserID),
		RootOrganizationID:      tenantapi.UUID(rootOrganizationID),
		ProtectedAdminBindingID: tenantapi.UUID(protectedAdminID),
		Kind:                    tenantapi.OnboardingKind(operationKind),
		DeploymentMode:          tenantapi.DeploymentMode(deploymentMode),
		BindingPurpose:          tenantapi.TenantBindingPurpose(bindingPurpose), SchemaName: schemaName,
		IdempotencyKey: idempotencyKey, RequestSHA256: requestSHA,
		ManifestSHA256: manifestSHA, Offering: tenantapi.BaseOfferingPin{
			VersionID: tenantapi.UUID(offeringVersionID), Revision: offeringRevision,
			SHA256: offeringSHA, EffectiveAt: offeringEffectiveAt.UTC()},
		Authority: tenantapi.ProtectedAuthorityPlan{
			RegistryBindingID: tenantapi.UUID(registryBindingID),
			ScopeKind:         tenantapi.AuthorityScopeKind(authorityKind),
			ScopeID:           tenantapi.UUID(authorityScopeID), Generation: authorityGeneration,
			StateSHA256: authorityState, ProposalSHA256: authorityProposal},
		CreatedAt: createdAt.UTC(),
	})
	if err != nil || plan.StepPlanSHA256() != stepPlanSHA {
		return workerapp.Claim{}, workerapp.ErrInvalid
	}
	return workerapp.Claim{Plan: plan, Customer: workerapp.CustomerSnapshot{
		Name: customerName, Slug: tenantSlug,
		FirstAdmin: workerapp.FirstAdminSnapshot{DisplayName: adminDisplayName, Email: adminEmail}},
		State: tenantapi.OnboardingState(state), Revision: revision,
		LeaseTokenSHA256: leaseTokenSHA256, LeaseUntil: leaseUntil.UTC()}, nil
}
