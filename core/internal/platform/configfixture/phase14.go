package configfixture

import "path/filepath"

const fixtureAuthorityKMSKeyID = "gcp-kms://projects/harpy-fixture/locations/europe-west1/" +
	"keyRings/authority/cryptoKeys/proposer/cryptoKeyVersions/7"

func setPhase14API(setter interface {
	EnvironmentSetter
	ArtifactFixture
}) {
	for _, item := range phase14APIEntries(setter) {
		setter.Setenv(item.name, item.value)
	}
}

func setPhase14Worker(setter interface {
	EnvironmentSetter
	ArtifactFixture
}) {
	for _, item := range phase14WorkerEntries(setter) {
		setter.Setenv(item.name, item.value)
	}
}

func phase14APIEnvironment(fixture ArtifactFixture) []string {
	return environmentStrings(phase14APIEntries(fixture))
}

func phase14WorkerEnvironment(fixture ArtifactFixture) []string {
	return environmentStrings(phase14WorkerEntries(fixture))
}

func phase14APIEntries(fixture ArtifactFixture) []entry {
	fixture.Helper()
	directory := fixture.TempDir()
	ackTrust := phase14Artifact(fixture, directory, "authority-ack-trust.json")
	return []entry{
		{"HARPY_SAAS_AUTHORITY_PROPOSER_KMS_KEY_ID", fixtureAuthorityKMSKeyID},
		{"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH", ackTrust},
	}
}

func phase14WorkerEntries(fixture ArtifactFixture) []entry {
	fixture.Helper()
	directory := fixture.TempDir()
	return []entry{
		{"HARPY_LICENSE_SNAPSHOT_FILE", phase14Artifact(fixture, directory, "license.json")},
		{"HARPY_LICENSE_TRUST_BUNDLE_PATH", phase14Artifact(fixture, directory, "license-trust.json")},
		{"HARPY_RECOVERY_TRUST_BUNDLE_PATH", phase14Artifact(fixture, directory, "recovery-trust.json")},
		{"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH", phase14Artifact(fixture, directory, "authority-ack-trust.json")},
		{"HARPY_RELEASE_TRUST_ROOT_PATH", phase14Artifact(fixture, directory, "release-trust.json")},
	}
}

func phase14Artifact(
	fixture ArtifactFixture,
	directory string,
	name string,
) string {
	path := filepath.Join(directory, name)
	if err := writeFixture(path, []byte("fixture\n")); err != nil {
		fixture.Fatalf("create Phase 14 config fixture: %v", err)
	}
	return path
}
