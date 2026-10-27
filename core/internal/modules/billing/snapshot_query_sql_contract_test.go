package billing

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var billingPlatformReference = regexp.MustCompile(
	`platform\.(billing_[a-z0-9_]+)`,
)

func TestBillingSnapshotQueryAdaptersUseOnlyGrantedReadSurfaces(t *testing.T) {
	root := filepath.Join("adapters", "postgres")
	wantPackages := []string{
		"catalogquery",
		"discounteligibility",
		"subscription",
		"usage",
	}
	if got := billingSnapshotQueryPackages(t, root); !slices.Equal(got, wantPackages) {
		t.Fatalf("billing SnapshotQuery packages = %v, want %v", got, wantPackages)
	}

	manifests := map[string]string{
		"catalogquery":        "sql.go",
		"discounteligibility": "sql.go",
		"subscription":        "query_sql.go",
		"usage":               "query_sql.go",
	}
	allowed := map[string]bool{
		"billing_catalog_campaign_snapshot_read_v1":     true,
		"billing_catalog_offering_snapshot_read_v1":     true,
		"billing_discount_eligibility_snapshot_read_v1": true,
		"billing_discount_offerings_snapshot_read_v1":   true,
		"billing_legacy_identity_alias_v1":              true,
		"billing_promotion_discount_read_v2":            true,
		"billing_resolve_current_global_authority_v1":   true,
		"billing_resolve_current_tenant_authority_v1":   true,
		"billing_subscription_snapshot_read_v1":         true,
		"billing_usage_context_snapshot_read_v1":        true,
		"billing_usage_snapshot_read_v1":                true,
	}
	for packageName, manifestName := range manifests {
		manifestPath := filepath.Join(root, packageName, manifestName)
		manifest := readBillingContractFile(t, manifestPath)
		for _, embeddedPath := range billingEmbeddedSQLPaths(manifest) {
			sqlPath := filepath.Join(root, packageName, filepath.FromSlash(embeddedPath))
			source := readBillingContractFile(t, sqlPath)
			for _, match := range billingPlatformReference.FindAllStringSubmatch(source, -1) {
				if !allowed[match[1]] {
					t.Errorf("%s directly references unapproved Billing surface %s",
						sqlPath, match[0])
				}
			}
		}
	}
}

func TestBillingSnapshotQueryManifestsExcludeCommandLocks(t *testing.T) {
	for _, path := range []string{
		filepath.Join("adapters", "postgres", "subscription", "query_sql.go"),
		filepath.Join("adapters", "postgres", "usage", "query_sql.go"),
	} {
		source := readBillingContractFile(t, path)
		for _, forbidden := range []string{
			"resolve_period.sql", "lock_snapshot.sql", "lock_subscription.sql",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s embeds command-only SQL %q", path, forbidden)
			}
		}
	}
}

func billingSnapshotQueryPackages(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var packages []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files, err := filepath.Glob(filepath.Join(root, entry.Name(), "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, path := range files {
			if strings.Contains(readBillingContractFile(t, path),
				"platformtx.SnapshotQuery") {
				found = true
				break
			}
		}
		if found {
			packages = append(packages, entry.Name())
		}
	}
	slices.Sort(packages)
	return packages
}

func billingEmbeddedSQLPaths(manifest string) []string {
	var paths []string
	for _, line := range strings.Split(manifest, "\n") {
		const marker = "//go:embed "
		if index := strings.Index(line, marker); index >= 0 {
			paths = append(paths, strings.TrimSpace(line[index+len(marker):]))
		}
	}
	return paths
}

func readBillingContractFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
