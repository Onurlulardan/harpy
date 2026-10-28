package main

import (
	"fmt"
	"strings"
)

func localAPIAuthorityEnvironment() profileEnvironment {
	return profileEnvironment{
		"HARPY_RECOVERY_TRUST_BUNDLE_PATH": "/run/recovery-signing/current/" +
			"recovery-trust-bundle.v1.json",
		"HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE": "/run/recovery-signing/current/" +
			"recovery-control-signer.ed25519",
		"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH": "/run/dev-authority/" +
			"current/protected-authority-ack-trust-bundle.v1.json",
	}
}

func localWorkerAuthorityEnvironment() profileEnvironment {
	return profileEnvironment{
		"HARPY_LICENSE_SNAPSHOT_FILE": "/run/dev-authority/current/" +
			"license-snapshot.json",
		"HARPY_LICENSE_TRUST_BUNDLE_PATH": "/run/dev-authority/current/" +
			"license-trust.json",
		"HARPY_RECOVERY_TRUST_BUNDLE_PATH": "/run/dev-authority/current/" +
			"recovery-trust-bundle.v1.json",
		"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH": "/run/dev-authority/" +
			"current/protected-authority-ack-trust-bundle.v1.json",
		"HARPY_RELEASE_TRUST_ROOT_PATH": "/run/dev-authority/current/" +
			"protected-authority-release-trust-root.json",
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH": "/run/dev-authority/" +
			"protected-authority-heads.journal",
		"HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE": secretRoot +
			"local-authority-ack-signing-seed",
	}
}

func refreshAuthorityProfileEnvironments(set *artifactSet) error {
	for name, updates := range map[string]profileEnvironment{
		"api.env":    localAPIAuthorityEnvironment(),
		"worker.env": localWorkerAuthorityEnvironment(),
	} {
		values, err := mergeProfileEnvironment(set.files[name], updates)
		if err != nil {
			return fmt.Errorf("refresh %s: %w", name, err)
		}
		set.files[name] = values
	}
	return nil
}

func mergeProfileEnvironment(
	document []byte,
	updates profileEnvironment,
) ([]byte, error) {
	values := make(profileEnvironment)
	for _, line := range strings.Split(strings.TrimSpace(string(document)), "\n") {
		key, value, found := strings.Cut(line, "=")
		if !found || !validEnvironmentKey(key) || value == "" {
			return nil, fmt.Errorf("source profile is invalid")
		}
		if _, duplicate := values[key]; duplicate {
			return nil, fmt.Errorf("source profile contains duplicate key")
		}
		values[key] = value
	}
	for key, value := range updates {
		values[key] = value
	}
	return renderProfileEnvironment(values)
}
