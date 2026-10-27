package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func validIAMSystemTransferEnvironment() []string {
	directory, err := os.MkdirTemp("", "harpy-iam-transfer-config-")
	if err != nil {
		panic(err)
	}
	spool := filepath.Join(directory, "spool")
	if err := os.Mkdir(spool, 0o700); err != nil {
		panic(err)
	}
	exportKeyring := filepath.Join(directory, "export-keyring.json")
	importKeyring := filepath.Join(directory, "import-keyring.json")
	writeFixtureKeyring(exportKeyring, "iam-export-2026-07", 111)
	writeFixtureKeyring(importKeyring, "iam-import-2026-07", 121)
	return []string{iamTransferSpoolKey + "=" + spool,
		iamExportEndpointKey + "=https://s3.fixture.invalid",
		iamExportRegionKey + "=eu-central-1",
		iamExportBucketKey + "=harpy-iam-system-exports",
		iamExportOwnerKey + "=000000000000", iamExportPathKey + "=true",
		iamExportSourceKey + "=workload_identity", iamExportKeyringKey + "=" + exportKeyring,
		iamImportEndpointKey + "=https://s3.fixture.invalid",
		iamImportRegionKey + "=eu-central-1",
		iamImportBucketKey + "=harpy-iam-system-import-sources",
		iamImportOwnerKey + "=000000000000", iamImportPathKey + "=true",
		iamImportSourceKey + "=workload_identity", iamImportKeyringKey + "=" + importKeyring}
}

func writeFixtureKeyring(path, id string, marker byte) {
	material := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 32))
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":%q,"key":%q},"retained":[]}`,
		id, material)
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		panic(err)
	}
}

func TestIAMSystemTransferArtifactsAreSeparatedAndRedacted(t *testing.T) {
	loaded, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	settings, err := loaded.IAMSystemTransferArtifacts()
	if err != nil || settings.Validate() != nil || fmt.Sprintf("%v", settings) !=
		"[REDACTED IAM ARTIFACTS]" {
		t.Fatalf("IAM transfer settings invalid or exposed: %v", err)
	}
}

func TestIAMSystemTransferRejectsSharedBucket(t *testing.T) {
	entries := withEnvironment(validAPIEnvironment(),
		iamImportBucketKey+"=harpy-iam-system-exports")
	if _, err := loadAPI(entries); err == nil {
		t.Fatal("shared IAM transfer bucket was accepted")
	}
}
