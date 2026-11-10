package billing

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOnPremBaseOfferingAllowsAbsentLicenseSource(t *testing.T) {
	input := onPremOfferingInput()
	if !validInstallationBaseOffering(input) {
		t.Fatal("on-prem offering rejected without a license source")
	}
	input.Source = InstallationOfferingSource{Present: true, Sequence: 1,
		SHA256: strings.Repeat("ab", 32)}
	if !validInstallationBaseOffering(input) {
		t.Fatal("explicit offering source was rejected")
	}
	input.Source = InstallationOfferingSource{Present: true}
	if validInstallationBaseOffering(input) {
		t.Fatal("blank license source was accepted")
	}
}

func onPremOfferingInput() InstallationBaseOfferingInput {
	return InstallationBaseOfferingInput{
		OperationID:       installationUUID("019b0000-0000-7000-8000-000000000001"),
		EnvironmentID:     installationUUID("019b0000-0000-7000-8000-000000000002"),
		TenantID:          installationUUID("019b0000-0000-7000-8000-000000000003"),
		OfferingVersionID: installationUUID("019b0000-0000-7000-8000-000000000004"),
		DeploymentMode:    "onpremise", BindingPurpose: "onpremise_primary",
		SchemaName:       "tenant_00112233445566778899aabbccddeeff",
		OfferingRevision: 1, OfferingSHA256: strings.Repeat("cd", 32),
		OfferingEffectiveAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		ManifestSHA256:      strings.Repeat("ef", 32),
	}
}

func installationUUID(value string) [16]byte {
	return [16]byte(uuid.MustParse(value))
}
