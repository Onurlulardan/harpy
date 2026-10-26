package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func validAnalyticsDashboardExportEnvironment(worker bool) []string {
	directory, err := os.MkdirTemp("", "harpy-analytics-export-config-")
	if err != nil {
		panic(err)
	}
	keyringPath := filepath.Join(directory, "fcart-kek-keyring.json")
	material := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{141}, 32))
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":"analytics-export-2026-08","key":"%s"},"retained":[]}`,
		material,
	)
	if err = os.WriteFile(keyringPath, []byte(document), 0o600); err != nil {
		panic(err)
	}
	entries := []string{
		analyticsExportEndpointKey + "=https://s3.fixture.invalid",
		analyticsExportRegionKey + "=eu-central-1",
		analyticsExportBucketKey + "=harpy-analytics-dashboard-exports",
		analyticsExportOwnerKey + "=000000000000",
		analyticsExportPathKey + "=true",
		analyticsExportSourceKey + "=workload_identity",
		analyticsExportKeyringKey + "=" + keyringPath,
	}
	if worker {
		entries = append(entries, analyticsExportSpoolKey+"="+directory)
	}
	return entries
}
