package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func validFormsExportEnvironment(worker bool) []string {
	directory, err := os.MkdirTemp("", "harpy-forms-export-config-")
	if err != nil {
		panic(err)
	}
	keyringPath := filepath.Join(directory, "fcart-kek-keyring.json")
	if err := os.WriteFile(keyringPath, []byte(validFormsExportKeyringJSON()), 0o600); err != nil {
		panic(err)
	}
	entries := []string{
		formsExportEndpointKey + "=https://s3.fixture.invalid",
		formsExportRegionKey + "=eu-central-1",
		formsExportBucketKey + "=harpy-form-exports",
		formsExportOwnerKey + "=000000000000",
		formsExportPathStyleKey + "=true",
		formsExportCredentialSourceKey + "=workload_identity",
		formsExportFCARTKeyringFileKey + "=" + keyringPath,
	}
	if worker {
		entries = append(entries, "HARPY_FORMS_EXPORT_SPOOL_DIR="+directory)
	}
	return entries
}

func validFormsExportKeyringJSON() string {
	material := func(marker byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 32))
	}
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":"forms-export-2026-07","key":"%s"},"retained":[]}`,
		material(71),
	)
}
