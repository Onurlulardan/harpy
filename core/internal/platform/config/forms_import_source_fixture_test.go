package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func validFormsImportSourceEnvironment(api bool) []string {
	directory, err := os.MkdirTemp("", "harpy-forms-import-source-config-")
	if err != nil {
		panic(err)
	}
	keyringPath := filepath.Join(directory, "fcart-kek-keyring.json")
	keyring := fmt.Sprintf(
		`{"version":1,"current":{"kid":"forms-import-source-2026-07","key":"%s"},"retained":[]}`,
		base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{91}, 32)))
	if err := os.WriteFile(keyringPath, []byte(keyring), 0o600); err != nil {
		panic(err)
	}
	entries := []string{formsImportSourceEndpointKey + "=https://s3.fixture.invalid",
		formsImportSourceRegionKey + "=eu-central-1",
		formsImportSourceBucketKey + "=harpy-form-import-sources",
		formsImportSourceOwnerKey + "=000000000000",
		formsImportSourcePathStyleKey + "=true",
		formsImportSourceCredentialSourceKey + "=workload_identity",
		formsImportSourceFCARTKeyringFileKey + "=" + keyringPath}
	if api {
		staging := filepath.Join(directory, "staging")
		if err := os.Mkdir(staging, 0o700); err != nil {
			panic(err)
		}
		entries = append(entries, formsImportSourceStagingKey+"="+staging)
	}
	return entries
}
