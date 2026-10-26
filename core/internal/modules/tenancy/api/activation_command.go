package api

import tenantactivationv1 "github.com/Onurlulardan/harpy/contracts/tenantactivation/v1"

type ActivationCommand = tenantactivationv1.Command

const (
	ActivateTenantCommand            = tenantactivationv1.ActivateTenant
	InitializeTenantIdentityCommand  = tenantactivationv1.InitializeTenantIdentity
	AssignTenantBaseOfferingCommand  = tenantactivationv1.AssignTenantBaseOffering
	OpenTenantUsageCommand           = tenantactivationv1.OpenTenantUsage
	InitializeSystemBaselineCommand  = tenantactivationv1.InitializeSystemBaseline
	InitializeRecoverySigningCommand = tenantactivationv1.InitializeRecoverySigning
	BindInstanceOwnerCommand         = tenantactivationv1.BindInstanceOwner
	PinProtectedAuthorityHeadCommand = tenantactivationv1.PinProtectedAuthorityHead
	CompleteTenantActivationCommand  = tenantactivationv1.CompleteTenantActivation
)

func ActivationCommandManifest(kind OnboardingKind) []ActivationCommand {
	return tenantactivationv1.Manifest(string(kind))
}

func activationManifestText(kind OnboardingKind) string {
	return tenantactivationv1.Text(string(kind))
}
