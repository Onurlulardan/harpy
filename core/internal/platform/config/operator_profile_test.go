package config

import (
	"strings"
	"testing"
)

func TestOperatorProfilesLoadExactArtifacts(t *testing.T) {
	initialize, err := loadOperatorInitialize(validOperatorInitializeEnvironment(DeploymentOnPrem))
	if err != nil {
		t.Fatal(err)
	}
	if initialize.InitialPassword().Access() != ArtifactReadOnceCleanup ||
		initialize.RecoverySigningDirectory().Access() != ArtifactCreateOnly {
		t.Fatal("initialize artifact access contract drifted")
	}
	if _, err := loadOperatorInitialize(validOperatorInitializeEnvironment(DeploymentSaaS)); err != nil {
		t.Fatalf("SaaS initialize rejected: %v", err)
	}
	identity, err := loadOperatorIdentityReconcile(validOperatorIdentityEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if identity.IAMCustodyManifest().Access() != ArtifactReadOnly {
		t.Fatal("identity reconciler custody mount is not read-only")
	}
	license, err := loadOperatorLicenseReconcile(validOperatorLicenseEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if license.LicenseSource().Kind() != LicenseSourceSnapshot {
		t.Fatal("license reconciler source branch drifted")
	}
}

func TestOperatorProfilesRejectCrossProfileAndModeKeys(t *testing.T) {
	identity := append(validOperatorIdentityEnvironment(),
		licenseSnapshotKey+"=/must-not-be-visible")
	if _, err := loadOperatorIdentityReconcile(identity); err == nil {
		t.Fatal("identity reconciler accepted license source")
	}
	license := append(validOperatorLicenseEnvironment(), deploymentModeEnvironmentKey+"=saas")
	if _, err := loadOperatorLicenseReconcile(license); err == nil {
		t.Fatal("SaaS license reconciler branch was accepted")
	}
	onPrem := append(validOperatorInitializeEnvironment(DeploymentOnPrem),
		saasRegistryURLKey+"=https://authority.invalid")
	if _, err := loadOperatorInitialize(onPrem); err == nil {
		t.Fatal("on-premise initialize accepted SaaS registry input")
	}
}

func TestOnPremOperatorInitializeSkipsSharedAckBundleAndReleaseTrust(t *testing.T) {
	directory := phase14FixtureDirectory("shared-onprem-operator")
	shared := append(validOperatorInitializeEnvironment(DeploymentOnPrem),
		ackTrustBundleKey+"="+phase14FixtureFile(directory, "ack-trust.json", false),
		"HARPY_RELEASE_TRUST_ROOT_PATH="+phase14FixtureFile(directory, "release-trust.json", false),
	)
	if _, err := loadOperatorInitialize(shared); err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("on-prem operator initialize accepted shared authority files: %v", err)
	}
	loaded, err := loadOperatorInitialize(entriesAllowedForMode(shared, "operator_initialize", string(DeploymentOnPrem)))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DeploymentMode() != DeploymentOnPrem ||
		loaded.AckTrustBundle().Path().Value() != "" ||
		loaded.ReleaseTrustRoot().Path().Value() != "" {
		t.Fatal("on-prem operator initialize consumed shared authority files")
	}
	if !StartupProfileAllowsForMode("operator_initialize", ackTrustBundleKey, string(DeploymentSaaS)) ||
		!StartupProfileAllows("api", ackTrustBundleKey) ||
		!StartupProfileAllows("worker", "HARPY_RELEASE_TRUST_ROOT_PATH") {
		t.Fatal("SaaS operator, API, or worker lost a required authority file")
	}
	api := withoutEnvironmentKey(validAPIEnvironment(), ackTrustBundleKey)
	if _, err := loadAPI(api); err == nil || !strings.Contains(err.Error(), ackTrustBundleKey) {
		t.Fatalf("API accepted a missing ack trust bundle: %v", err)
	}
}

func entriesAllowedForMode(entries []string, profileName, mode string) []string {
	allowed := make([]string, 0, len(entries))
	for _, entry := range entries {
		key, _, found := strings.Cut(entry, "=")
		if found && !StartupProfileAllowsForMode(profileName, key, mode) {
			continue
		}
		allowed = append(allowed, entry)
	}
	return allowed
}

func TestTenantMigrationExecutorOwnsModeAndReleaseInputs(t *testing.T) {
	loaded, err := loadTenantMigrationExecutor(validTenantMigrationEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DeploymentMode() != DeploymentOnPrem ||
		loaded.Release().ManifestPath().Value() == "" {
		t.Fatal("tenant migration profile contract drifted")
	}
}
