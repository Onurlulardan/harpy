package main

import (
	"encoding/json"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestExtensionReceiptUsesManifestControlHashes(t *testing.T) {
	files, err := buildReleaseFiles(newReleaseClock(time.Date(2026, 7, 22, 8, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	manifestBody := releaseBody(t, files, "extension-manifest.json")
	receiptBody := releaseBody(t, files, "extension-receipt.json")
	var manifest postgresv1.ExtensionManifest
	var receipt struct {
		Payload runtimev1.ExtensionReceiptPayload `json:"payload"`
	}
	if err = attestation.DecodeStrict(manifestBody, &manifest); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(receiptBody, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Extensions) != len(receipt.Payload.Extensions) {
		t.Fatalf("extension count = %d", len(receipt.Payload.Extensions))
	}
	for index, extension := range manifest.Extensions {
		record := receipt.Payload.Extensions[index]
		if record.Name != extension.Name || record.ControlFileSHA256 != extension.ControlFileSHA256 {
			t.Fatalf("receipt control hash for %s differs from the manifest", extension.Name)
		}
	}
}

func releaseBody(t *testing.T, files []releaseFile, name string) []byte {
	t.Helper()
	for _, file := range files {
		if file.name == name {
			return file.body
		}
	}
	t.Fatalf("missing %s", name)
	return nil
}
