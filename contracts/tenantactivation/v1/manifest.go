package v1

type Command string

const (
	ActivateTenant            Command = "tenancy.activate_tenant"
	InitializeTenantIdentity  Command = "iam.initialize_tenant_identity"
	AssignTenantBaseOffering  Command = "billing.assign_tenant_base_offering"
	OpenTenantUsage           Command = "billing.open_tenant_usage"
	InitializeSystemBaseline  Command = "systemresources.initialize_tenant_baseline"
	InitializeRecoverySigning Command = "platformops.initialize_recovery_signing_identity"
	BindInstanceOwner         Command = "platformauthority.bind_owner"
	PinProtectedAuthorityHead Command = "platformauthority.pin_protected_authority_head"
	CompleteTenantActivation  Command = "activation.complete"
)

var environmentCommands = []Command{
	ActivateTenant, InitializeTenantIdentity, AssignTenantBaseOffering,
	OpenTenantUsage, InitializeSystemBaseline, InitializeRecoverySigning,
	BindInstanceOwner, PinProtectedAuthorityHead, CompleteTenantActivation,
}

var customerCommands = []Command{
	ActivateTenant, InitializeTenantIdentity, AssignTenantBaseOffering,
	OpenTenantUsage, InitializeSystemBaseline,
	PinProtectedAuthorityHead, CompleteTenantActivation,
}

func Manifest(kind string) []Command {
	if kind == "saas_customer_tenant_provision" {
		return append([]Command(nil), customerCommands...)
	}
	return append([]Command(nil), environmentCommands...)
}

func Text(kind string) string {
	manifest := Manifest(kind)
	result := ""
	for index, command := range manifest {
		if index > 0 {
			result += "\n"
		}
		result += string(command)
	}
	return result
}
