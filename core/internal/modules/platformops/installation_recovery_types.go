package platformops

import (
	"context"
	"regexp"
	"strings"
)

var installationDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type InstallationActivationRow interface{ Scan(...any) error }

type InstallationActivationExecutor struct {
	QueryRow func(context.Context, string, ...any) InstallationActivationRow
}

type InstallationRecoveryIdentityInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	SchemaName                           string
	DeploymentMode                       string
	SigningGeneration                    int64
	TrustBundleSHA256                    string
	RecoverySigningPresent               bool
	ManifestSHA256                       string
}

func validInstallationRecoveryIdentity(input InstallationRecoveryIdentityInput) bool {
	if input.OperationID == [16]byte{} || input.EnvironmentID == [16]byte{} ||
		input.TenantID == [16]byte{} || len(input.SchemaName) < 3 ||
		len(input.SchemaName) > 63 || !strings.HasPrefix(input.SchemaName, "t_") ||
		!installationDigestPattern.MatchString(input.ManifestSHA256) {
		return false
	}
	switch input.DeploymentMode {
	case "onpremise":
		return input.RecoverySigningPresent && input.SigningGeneration > 0 &&
			installationDigestPattern.MatchString(input.TrustBundleSHA256)
	case "saas":
		return !input.RecoverySigningPresent && input.SigningGeneration == 0 &&
			input.TrustBundleSHA256 == ""
	default:
		return false
	}
}
