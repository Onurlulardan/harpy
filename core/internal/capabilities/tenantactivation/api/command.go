package api

import tenantactivationv1 "github.com/Onurlulardan/harpy/contracts/tenantactivation/v1"

type Command = tenantactivationv1.Command

const (
	ActivateTenant            = tenantactivationv1.ActivateTenant
	InitializeTenantIdentity  = tenantactivationv1.InitializeTenantIdentity
	AssignTenantBaseOffering  = tenantactivationv1.AssignTenantBaseOffering
	OpenTenantUsage           = tenantactivationv1.OpenTenantUsage
	InitializeSystemBaseline  = tenantactivationv1.InitializeSystemBaseline
	InitializeRecoverySigning = tenantactivationv1.InitializeRecoverySigning
	BindInstanceOwner         = tenantactivationv1.BindInstanceOwner
	PinProtectedAuthorityHead = tenantactivationv1.PinProtectedAuthorityHead
	CompleteTenantActivation  = tenantactivationv1.CompleteTenantActivation
)

func commands(kind string) []Command { return tenantactivationv1.Manifest(kind) }

func Manifest(kind string) []Command { return tenantactivationv1.Manifest(kind) }
