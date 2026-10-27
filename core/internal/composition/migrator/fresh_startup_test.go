package migrator

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func TestLoadStartupRegistryUsesFreshDocuments(t *testing.T) {
	documents, err := deployMigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	registry, err := migration.LoadStartupRegistry(documents)
	if err != nil || registry == nil {
		t.Fatalf("startup registry = %#v, %v", registry, err)
	}
	paths := registry.FreshPaths()
	if registry.FreshEntryCount() != len(documents) || len(paths) != len(documents) ||
		len(registry.Digests()) != 1 {
		t.Fatalf("fresh entries = %d documents = %d digests = %d",
			registry.FreshEntryCount(), len(documents), len(registry.Digests()))
	}
	if len(paths) != 2071 {
		t.Fatalf("selected fresh entry count = %d, want 2071", len(paths))
	}
	versions := make(map[int64]string, len(paths))
	owners := map[string]bool{}
	for _, documentPath := range paths {
		if strings.Contains(documentPath, "/v1_adopted/") ||
			strings.HasSuffix(documentPath, ".v1_adopted.sql") {
			t.Fatalf("adopted path entered startup registry: %s", documentPath)
		}
		version, parseErr := freshStartupVersion(documentPath)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if previous, exists := versions[version]; exists {
			t.Fatalf("version %d applies twice: %s and %s", version, previous, documentPath)
		}
		versions[version] = documentPath
		owners[freshStartupOwner(documentPath)] = true
	}
	assertFreshTerminal(t, versions)
	for _, owner := range []string{"audit", "usage", "support"} {
		if !owners[owner] {
			t.Fatalf("owner %s is absent from the fresh startup set", owner)
		}
	}
	encoded, err := schemacontract.EncodeFreshBinaryContract(registry)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil || !bytes.Equal(contract, encoded) {
		t.Fatalf("embedded fresh contract differs from computed manifest hash")
	}
	embedded, loaded, err := schemacontract.LoadStartup(documents)
	if err != nil || loaded == nil || !embedded.IsVerified() ||
		embedded.Value().MigrationManifestSHA256 != registry.Digests()[0] {
		t.Fatalf("load startup = %#v, %v", loaded, err)
	}
}

func assertFreshTerminal(t *testing.T, versions map[int64]string) {
	t.Helper()
	checks := map[int64]string{
		20260730120104: "/supersessions/v3/fresh/",
		20260729100025: "/iam/migrations/tenant/supersessions/fresh/",
		20260802234002: "/iam/migrations/tenant/supersessions/fresh/",
		20260802010000: "/analytics/migrations/tenant/supersessions/fresh/",
		20260728105100: "/catalog/migrations/tenant/supersessions/fresh/",
		20260727090000: "/tenancy/migrations/platform/supersessions/fresh/",
		20260727090100: "/tenancy/migrations/platform/supersessions/fresh/",
		20260718020030: "/platform/async/migrations/supersessions/fresh/",
	}
	for version, fragment := range checks {
		if !strings.Contains(versions[version], fragment) {
			t.Fatalf("version %d path = %s", version, versions[version])
		}
	}
}

func freshStartupVersion(documentPath string) (int64, error) {
	base := documentPath[strings.LastIndex(documentPath, "/")+1:]
	prefix, _, _ := strings.Cut(base, "_")
	return strconv.ParseInt(prefix, 10, 64)
}

func freshStartupOwner(documentPath string) string {
	const moduleRoot = "core/internal/modules/"
	if strings.Contains(documentPath, moduleRoot) {
		rest := documentPath[strings.Index(documentPath, moduleRoot)+len(moduleRoot):]
		return rest[:strings.Index(rest, "/")]
	}
	const platformRoot = "core/internal/platform/"
	rest := documentPath[strings.Index(documentPath, platformRoot)+len(platformRoot):]
	return "platform_" + rest[:strings.Index(rest, "/")]
}
