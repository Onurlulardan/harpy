package config

import (
	"strings"
	"testing"
)

func TestAPIAuthorityLoadsExactModeBranches(t *testing.T) {
	saas, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if saas.Authority().Mode() != DeploymentSaaS ||
		saas.Authority().AckTrustBundle().Access() != ArtifactReadOnly {
		t.Fatal("SaaS API authority branch drifted")
	}
	onPremEntries := withoutPhase14APIKeys(validAPIEnvironment())
	onPremEntries = append(onPremEntries, validAPIOnPremAuthorityEnvironment()...)
	onPrem, err := loadAPI(onPremEntries)
	if err != nil {
		t.Fatal(err)
	}
	if onPrem.Authority().Mode() != DeploymentOnPrem ||
		onPrem.Authority().RecoveryControlKey().Access() != ArtifactReadOnly {
		t.Fatal("on-premise API authority branch drifted")
	}
}

func TestAPIAuthorityRejectsModeLeakageAndInvalidKMSID(t *testing.T) {
	tests := [][]string{
		append(validAPIEnvironment(),
			recoveryTrustBundleKey+"="+phase14FixtureFile(
				phase14FixtureDirectory("api-leak"), "recovery.json", false)),
		withEnvironment(validAPIEnvironment(), saasKMSKeyIDKey+"=gcp-kms://alias/current"),
	}
	for _, entries := range tests {
		if _, err := loadAPI(entries); err == nil {
			t.Fatal("invalid API authority branch was accepted")
		}
	}
}

func TestWorkerAuthorityLoadsDevelopmentAndSaaSBranches(t *testing.T) {
	development, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if development.Authority().Mode() != WorkerAuthorityDevelopment ||
		development.Authority().LocalJournal().Access() != ArtifactReadWrite {
		t.Fatal("development worker authority branch drifted")
	}
	saasEntries := withoutPhase14WorkerKeys(validWorkerEnvironment())
	directory := phase14FixtureDirectory("worker-saas")
	saasEntries = append(saasEntries,
		deploymentModeEnvironmentKey+"=saas",
		ackTrustBundleKey+"="+phase14FixtureFile(directory, "ack-trust.json", false),
		saasRegistryURLKey+"=https://authority.invalid",
		"HARPY_RELEASE_TRUST_ROOT_PATH="+
			phase14FixtureFile(directory, "release-trust.json", false),
	)
	saas, err := loadWorker(saasEntries)
	if err != nil {
		t.Fatal(err)
	}
	if saas.Authority().Mode() != WorkerAuthoritySaaS ||
		saas.Authority().ReleaseTrustRoot().Access() != ArtifactReadOnly {
		t.Fatal("SaaS worker authority branch drifted")
	}
}

func TestWorkerAuthorityRejectsSourceAndModeLeakage(t *testing.T) {
	entries := append(validWorkerEnvironment(),
		licenseCredentialKey+"="+phase14FixtureFile(
			phase14FixtureDirectory("worker-source"), "credential", true),
		licenseIssuerURLKey+"=https://issuer.invalid",
	)
	if _, err := loadWorker(entries); err == nil || !strings.Contains(err.Error(), "LICENSE") {
		t.Fatalf("combined worker license sources accepted: %v", err)
	}
	saasLeak := withEnvironment(validWorkerEnvironment(), deploymentModeEnvironmentKey+"=saas")
	if _, err := loadWorker(saasLeak); err == nil {
		t.Fatal("SaaS worker accepted on-premise authority material")
	}
}

func TestSaaSWorkerAuthorityRequiresAbsoluteReleaseTrustRoot(t *testing.T) {
	directory := phase14FixtureDirectory("worker-release-root")
	entries := append(withoutPhase14WorkerKeys(validWorkerEnvironment()),
		deploymentModeEnvironmentKey+"=saas",
		ackTrustBundleKey+"="+phase14FixtureFile(directory, "ack.json", false),
		saasRegistryURLKey+"=https://authority.invalid",
	)
	if _, err := loadWorker(entries); err == nil ||
		!strings.Contains(err.Error(), "RELEASE_TRUST_ROOT_PATH") {
		t.Fatalf("missing SaaS worker release root accepted: %v", err)
	}
	entries = append(entries, "HARPY_RELEASE_TRUST_ROOT_PATH=relative/root.json")
	if _, err := loadWorker(entries); err == nil ||
		!strings.Contains(err.Error(), "RELEASE_TRUST_ROOT_PATH") {
		t.Fatalf("relative SaaS worker release root accepted: %v", err)
	}
}
