package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func validDocumentsEnvironment(worker bool) []string {
	directory, err := os.MkdirTemp("", "harpy-documents-config-")
	if err != nil {
		panic(err)
	}
	keyringPath := filepath.Join(directory, "documents-kek-keyring.json")
	if err := os.WriteFile(keyringPath, []byte(validDocumentsKeyringJSON()), 0o600); err != nil {
		panic(err)
	}
	entries := []string{
		documentsEndpointKey + "=https://documents-s3.fixture.invalid",
		documentsRegionKey + "=eu-central-1",
		documentsBucketKey + "=harpy-documents",
		documentsOwnerKey + "=000000000000",
		documentsPathKey + "=true",
		documentsSourceKey + "=workload_identity",
		documentsKEKKeyringFileKey + "=" + keyringPath,
		documentsBillingActorKey + "=019b1234-5678-7abc-8def-0123456789ab",
	}
	if !worker {
		return append(entries, documentsClamAVAddressKey+"=clamav.fixture.invalid:3310")
	}
	return append(entries,
		documentsClamAVAddressKey+"=clamav.fixture.invalid:3310",
		documentsLegacyEndpointKey+"=https://legacy-documents-s3.fixture.invalid",
		documentsLegacyRegionKey+"=eu-central-1",
		documentsLegacyBucketKey+"=harpy-documents-legacy",
		documentsLegacyOwnerKey+"=000000000000",
		documentsLegacyPathKey+"=true",
		documentsLegacySourceKey+"=workload_identity",
		documentsArchiveReservationKey+"=24h",
	)
}

func validDocumentsKeyringJSON() string {
	material := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{91}, 32))
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":"documents-2026-07","key":"%s"},"retained":[]}`,
		material,
	)
}
