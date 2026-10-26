package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func TestRunProjectsPurposeSeparatedRecoveryArtifacts(t *testing.T) {
	source, api, worker := fixture(t)
	arguments := []string{
		"-source", source, "-api-target", api, "-worker-target", worker,
	}
	if err := run(arguments); err != nil {
		t.Fatal(err)
	}
	if err := run(arguments); err != nil {
		t.Fatalf("idempotent projection failed: %v", err)
	}
	apiNames := names(t, api)
	workerNames := names(t, worker)
	if len(apiNames) != 2 || len(workerNames) != 1 ||
		workerNames[0] != recoveryv1.BundleFilename {
		t.Fatalf("unexpected projection: api=%v worker=%v", apiNames, workerNames)
	}
}

func TestRunRejectsMismatchedControlSigner(t *testing.T) {
	source, api, worker := fixture(t)
	if err := os.WriteFile(
		filepath.Join(source, recoveryv1.ControlSignerFilename),
		make([]byte, ed25519.SeedSize), 0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{
		"-source", source, "-api-target", api, "-worker-target", worker,
	}); err == nil {
		t.Fatal("mismatched signer accepted")
	}
}

func fixture(t *testing.T) (string, string, string) {
	t.Helper()
	source, api, worker := t.TempDir(), t.TempDir(), t.TempDir()
	purposes := []recoveryv1.Purpose{
		recoveryv1.PurposeRecoveryManifest,
		recoveryv1.PurposeRecoveryControl,
		recoveryv1.PurposeRestoreReceipt,
	}
	material := make([]recoveryv1.PublicMaterial, 0, len(purposes))
	var controlSeed []byte
	for _, purpose := range purposes {
		public, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		var encoded [ed25519.PublicKeySize]byte
		copy(encoded[:], public)
		material = append(material, recoveryv1.PublicMaterial{
			Purpose: purpose, PublicKey: encoded,
		})
		if purpose == recoveryv1.PurposeRecoveryControl {
			controlSeed = private.Seed()
		}
	}
	bundle, err := recoveryv1.Build(
		"019b1234-5678-7abc-8def-0123456789ab",
		time.Now().UTC().Truncate(time.Microsecond), material,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(source, recoveryv1.BundleFilename),
		bundle.CanonicalBytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(source, recoveryv1.ControlSignerFilename),
		controlSeed, 0o600); err != nil {
		t.Fatal(err)
	}
	return source, api, worker
}

func names(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	return result
}
