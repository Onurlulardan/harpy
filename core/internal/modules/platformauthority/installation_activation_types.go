package platformauthority

import (
	"context"
	"regexp"
	"strings"
)

var installationSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type InstallationActivationRow interface{ Scan(...any) error }

type InstallationActivationExecutor struct {
	QueryRow func(context.Context, string, ...any) InstallationActivationRow
}

type InstallationOwnerBindingInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	InstanceOwnerBindingID               [16]byte
	ProtectedAdminBindingID              [16]byte
	RegistryBindingID                    [16]byte
	SchemaName, DeploymentMode           string
	ManifestSHA256                       string
}

type InstallationHeadPinInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	RegistryBindingID, ScopeID           [16]byte
	SchemaName, ScopeKind                string
	Generation                           int64
	PreviousHeadSHA256                   string
	StateSHA256, ProposalSHA256          string
	AckSHA256, ManifestSHA256            string
}

func validInstallationCommon(
	operationID, environmentID, tenantID [16]byte,
	schemaName, manifestSHA256 string,
) bool {
	return operationID != [16]byte{} && environmentID != [16]byte{} &&
		tenantID != [16]byte{} && len(schemaName) >= 3 && len(schemaName) <= 63 &&
		strings.HasPrefix(schemaName, "t_") &&
		installationSHA256Pattern.MatchString(manifestSHA256)
}

func validOwnerBinding(input InstallationOwnerBindingInput) bool {
	return validInstallationCommon(input.OperationID, input.EnvironmentID,
		input.TenantID, input.SchemaName, input.ManifestSHA256) &&
		input.InstanceOwnerBindingID != [16]byte{} &&
		input.ProtectedAdminBindingID != [16]byte{} &&
		input.RegistryBindingID != [16]byte{} &&
		(input.DeploymentMode == "onpremise" || input.DeploymentMode == "saas")
}

func validHeadPin(input InstallationHeadPinInput) bool {
	return validInstallationCommon(input.OperationID, input.EnvironmentID,
		input.TenantID, input.SchemaName, input.ManifestSHA256) &&
		input.RegistryBindingID != [16]byte{} && validInstallationScope(input) &&
		input.Generation > 0 &&
		((input.Generation == 1 && input.PreviousHeadSHA256 == "") ||
			(input.Generation > 1 && installationSHA256Pattern.MatchString(
				input.PreviousHeadSHA256))) &&
		installationSHA256Pattern.MatchString(input.StateSHA256) &&
		installationSHA256Pattern.MatchString(input.ProposalSHA256) &&
		installationSHA256Pattern.MatchString(input.AckSHA256)
}

func validInstallationScope(input InstallationHeadPinInput) bool {
	return input.ScopeKind == "environment" && input.ScopeID == input.EnvironmentID ||
		input.ScopeKind == "tenant" && input.ScopeID == input.TenantID
}
