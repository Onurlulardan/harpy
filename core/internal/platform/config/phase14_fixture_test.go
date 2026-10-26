package config

import (
	"os"
	"path/filepath"
)

const validAuthorityKMSKeyID = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/authority/cryptoKeys/proposer/cryptoKeyVersions/7"

func validAPISaaSAuthorityEnvironment() []string {
	directory := phase14FixtureDirectory("api-saas")
	return []string{
		saasKMSKeyIDKey + "=" + validAuthorityKMSKeyID,
		ackTrustBundleKey + "=" + phase14FixtureFile(directory, "ack-trust.json", false),
	}
}

func validAPIOnPremAuthorityEnvironment() []string {
	directory := phase14FixtureDirectory("api-onprem")
	return []string{
		recoveryTrustBundleKey + "=" + phase14FixtureFile(directory, "recovery-trust.json", false),
		"HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE=" +
			phase14FixtureFile(directory, "recovery-control.key", true),
		ackTrustBundleKey + "=" + phase14FixtureFile(directory, "ack-trust.json", false),
	}
}

func validWorkerDevelopmentAuthorityEnvironment() []string {
	directory := phase14FixtureDirectory("worker-development")
	return []string{
		licenseSnapshotKey + "=" + phase14FixtureFile(directory, "license.json", false),
		licenseTrustBundleKey + "=" + phase14FixtureFile(directory, "license-trust.json", false),
		recoveryTrustBundleKey + "=" + phase14FixtureFile(directory, "recovery-trust.json", false),
		ackTrustBundleKey + "=" + phase14FixtureFile(directory, "ack-trust.json", false),
		"HARPY_RELEASE_TRUST_ROOT_PATH=" +
			phase14FixtureFile(directory, "release-trust.json", false),
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH=" +
			phase14FixtureFile(directory, "authority-head.jsonl", false),
		"HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE=" +
			phase14FixtureFile(directory, "authority-ack.key", true),
	}
}

func phase14FixtureDirectory(prefix string) string {
	directory, err := os.MkdirTemp("", "harpy-phase14-"+prefix+"-")
	if err != nil {
		panic(err)
	}
	return directory
}

func phase14FixtureFile(directory, name string, sensitive bool) string {
	return phase14FixtureFileContent(directory, name, "fixture\n", sensitive)
}

func phase14FixtureFileContent(
	directory string,
	name string,
	content string,
	sensitive bool,
) string {
	mode := os.FileMode(0o644)
	if sensitive {
		mode = 0o600
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		panic(err)
	}
	return path
}

func withoutPhase14APIKeys(entries []string) []string {
	keys := append(apiPhase14Keys(), deploymentModeEnvironmentKey)
	for _, key := range keys {
		entries = withoutEnvironmentKey(entries, key)
	}
	return entries
}

func withoutPhase14WorkerKeys(entries []string) []string {
	keys := append(workerPhase14Keys(), deploymentModeEnvironmentKey)
	for _, key := range keys {
		entries = withoutEnvironmentKey(entries, key)
	}
	return entries
}
