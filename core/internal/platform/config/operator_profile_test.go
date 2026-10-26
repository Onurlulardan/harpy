package config

import "testing"

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
