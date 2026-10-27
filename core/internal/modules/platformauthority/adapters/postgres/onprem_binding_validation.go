package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func validOnPremiseBinding(
	record ports.BindingRecord,
	proof *artifact.VerifiedBindingProof,
) bool {
	if record.DeploymentMode != artifact.ModeOnPremise || proof == nil ||
		record.RegistryKind != artifact.RegistryDevelopment ||
		record.LicenseBindingID == nil || record.FinalizationReceiptSHA256 == nil ||
		record.RecoveryTrustBundleSHA256 == nil || proof.Bytes == nil {
		return false
	}
	value := proof.Value
	return value.LicenseBindingID == record.BindingID &&
		*record.LicenseBindingID == record.BindingID &&
		value.EnvironmentID == record.EnvironmentID &&
		value.FinalizationReceiptSHA256 == *record.FinalizationReceiptSHA256 &&
		value.RecoveryTrustBundleSHA256 == *record.RecoveryTrustBundleSHA256 &&
		value.ControlKID == record.SignerKID
}
