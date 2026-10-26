package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInternalSaaSProfileIsExactKeylessAndIdempotent(t *testing.T) {
	output := filepath.Join(t.TempDir(), "profile")
	arguments := validArguments(output)
	var stdout, stderr bytes.Buffer
	if code := runCLI(arguments, &stdout, &stderr, time.Now); code != 0 {
		t.Fatalf("generate exit=%d stderr=%q", code, stderr.String())
	}
	assertExactSaaSProfile(t, output)
	manifestPath := filepath.Join(output, manifestName)
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCLI(arguments, &stdout, &stderr, time.Now); code != 0 ||
		!strings.Contains(stdout.String(), "current") {
		t.Fatalf("repeat exit=%d stdout=%q stderr=%q", code,
			stdout.String(), stderr.String())
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("manifest changed: %v", err)
	}
}

func TestInternalSaaSProfileRejectsDriftAndCustomerInputs(t *testing.T) {
	output := filepath.Join(t.TempDir(), "profile")
	var stdout, stderr bytes.Buffer
	if code := runCLI(validArguments(output), &stdout, &stderr, time.Now); code != 0 {
		t.Fatal(stderr.String())
	}
	changed := validArguments(output)
	changed[3] = "https://other-authority.example.test"
	stdout.Reset()
	stderr.Reset()
	if code := runCLI(changed, &stdout, &stderr, time.Now); code != 1 ||
		!strings.Contains(stderr.String(), "different or invalid") {
		t.Fatalf("drift exit=%d stderr=%q", code, stderr.String())
	}
	for _, flagName := range []string{
		"-license-snapshot-file", "-recovery-signing-key-file",
		"-deployment-mode",
	} {
		customer := append(validArguments(filepath.Join(t.TempDir(), "customer")),
			flagName, "forbidden")
		stderr.Reset()
		if code := runCLI(customer, &stdout, &stderr, time.Now); code != 2 ||
			!strings.Contains(stderr.String(), "flag provided but not defined") {
			t.Fatalf("%s exit=%d stderr=%q", flagName, code, stderr.String())
		}
	}
}

func TestInternalSaaSProfileRejectsUnversionedKMSBeforeWriting(t *testing.T) {
	for _, invalid := range []string{
		"gcp-kms://alias/current",
		"gcp-kms://projects/harpy-prod/locations/global/keyRings/authority/" +
			"cryptoKeys/proposer/cryptoKeyVersions/0",
		"gcp-kms://projects/harpy-prod/locations/global/keyRings/authority/" +
			"cryptoKeys/proposer/cryptoKeyVersions/current",
	} {
		output := filepath.Join(t.TempDir(), "profile")
		arguments := validArguments(output)
		arguments[5] = invalid
		var stdout, stderr bytes.Buffer
		if code := runCLI(arguments, &stdout, &stderr, time.Now); code != 1 {
			t.Fatalf("invalid KMS accepted: %q", invalid)
		}
		if _, err := os.Lstat(output); !os.IsNotExist(err) {
			t.Fatalf("invalid KMS wrote output: %v", err)
		}
	}
}

func assertExactSaaSProfile(t *testing.T, output string) {
	t.Helper()
	expected := map[string][]string{
		"api.saas.env":    {modeKey, proposerKey, ackTrustKey, wifKey},
		"worker.saas.env": {modeKey, registryKey, ackTrustKey, wifKey},
		"operator-initialize.saas.env": {
			modeKey, registryKey, proposerKey, ackTrustKey, wifKey,
		},
	}
	for name, keys := range expected {
		content, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
		if len(lines) != len(keys) {
			t.Fatalf("%s key count=%d want=%d:\n%s", name, len(lines), len(keys), text)
		}
		for _, key := range keys {
			if !strings.Contains(text, key+"=") {
				t.Fatalf("%s misses %s:\n%s", name, key, text)
			}
		}
		if strings.Count(text, modeKey+"=saas\n") != 1 ||
			strings.Count(text, modeKey+"=") != 1 {
			t.Fatalf("%s does not opt into SaaS", name)
		}
		for _, forbidden := range []string{
			"HARPY_LICENSE_", "HARPY_RECOVERY_",
			"HARPY_LOCAL_AUTHORITY_",
		} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s contains forbidden %s", name, forbidden)
			}
		}
	}
}

func validArguments(output string) []string {
	return []string{"-output", output,
		"-authority-registry-url", "https://authority.internal.example.test",
		"-proposer-kms-key-id", "gcp-kms://projects/harpy-prod/locations/" +
			"europe-west1/keyRings/authority/cryptoKeys/proposer/cryptoKeyVersions/7",
		"-ack-trust-bundle-path", "/run/secrets/authority-ack-trust.json",
		"-workload-identity-config-path", "/run/identity/workload.json"}
}
