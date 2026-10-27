package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"time"
)

var installationSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type InstallationActivationRow interface{ Scan(...any) error }

type InstallationActivationExecutor struct {
	QueryRow func(context.Context, string, ...any) InstallationActivationRow
}

type InstallationOfferingSource struct {
	Sequence int64
	SHA256   string
	Present  bool
}

type InstallationBaseOfferingInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	OfferingVersionID                    [16]byte
	DeploymentMode                       string
	BindingPurpose                       string
	SchemaName                           string
	OfferingRevision                     int64
	OfferingSHA256                       string
	OfferingEffectiveAt                  time.Time
	Source                               InstallationOfferingSource
	ManifestSHA256                       string
}

type InstallationUsageInput struct {
	OperationID, TenantID [16]byte
	ManifestSHA256        string
	SchemaName            string
}

func validInstallationBaseOffering(input InstallationBaseOfferingInput) bool {
	if zeroInstallationUUID(input.OperationID) || zeroInstallationUUID(input.EnvironmentID) ||
		zeroInstallationUUID(input.TenantID) || zeroInstallationUUID(input.OfferingVersionID) ||
		input.OfferingRevision < 1 || !installationSHA256Pattern.MatchString(input.OfferingSHA256) ||
		!installationSHA256Pattern.MatchString(input.ManifestSHA256) ||
		len(input.SchemaName) < 3 || len(input.SchemaName) > 63 ||
		input.OfferingEffectiveAt.IsZero() || input.OfferingEffectiveAt.Location() != time.UTC {
		return false
	}
	switch input.DeploymentMode {
	case "onpremise":
		return input.BindingPurpose == "onpremise_primary" &&
			input.Source.Present && input.Source.Sequence > 0 &&
			installationSHA256Pattern.MatchString(input.Source.SHA256)
	case "saas":
		return (input.BindingPurpose == "saas_control" ||
			input.BindingPurpose == "saas_customer") &&
			!input.Source.Present && input.Source.Sequence == 0 && input.Source.SHA256 == ""
	default:
		return false
	}
}

func billingInstallationID(operationID [16]byte, label string) [16]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy:billing:installation:v1"))
	_, _ = hash.Write(operationID[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(label))
	var result [16]byte
	copy(result[:], hash.Sum(nil))
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return result
}

func zeroInstallationUUID(value [16]byte) bool { return value == [16]byte{} }

func billingInstallationUUID(value [16]byte) string {
	raw := make([]byte, 32)
	hex.Encode(raw, value[:])
	return string(raw[:8]) + "-" + string(raw[8:12]) + "-" +
		string(raw[12:16]) + "-" + string(raw[16:20]) + "-" + string(raw[20:])
}
