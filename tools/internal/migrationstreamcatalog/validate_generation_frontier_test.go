package migrationstreamcatalog

import (
	"strconv"
	"strings"
	"testing"
)

func TestLoadRejectsTenantGenerationAbovePlatformFrontier(t *testing.T) {
	root := t.TempDir()
	platform := validStream("alpha", 10)
	platform.TargetDatabaseGeneration = 72
	tenant := validStream("zeta", 20)
	tenant.Scope = "tenant"
	tenant.TargetDatabaseGeneration = 73
	writeRoot(t, root, catalogVersion, []string{
		streamPath("alpha"),
		streamPath("zeta"),
	})
	writeOwner(t, root, ownerDocumentVersion, "alpha", []Stream{platform})
	writeOwner(t, root, ownerDocumentVersion, "zeta", []Stream{tenant})

	_, err := Load(root)
	if err == nil || !strings.Contains(
		err.Error(),
		"tenant migration generation frontier 73 exceeds platform frontier 72",
	) {
		t.Fatalf("Load error = %v", err)
	}
}

func TestLoadAcceptsTenantGenerationAtOrBelowPlatformFrontier(t *testing.T) {
	for _, tenantGeneration := range []int64{71, 72} {
		t.Run(strconv.FormatInt(tenantGeneration, 10), func(t *testing.T) {
			root := t.TempDir()
			platform := validStream("alpha", 10)
			platform.TargetDatabaseGeneration = 72
			tenant := validStream("zeta", 20)
			tenant.Scope = "tenant"
			tenant.TargetDatabaseGeneration = tenantGeneration
			writeRoot(t, root, catalogVersion, []string{
				streamPath("alpha"),
				streamPath("zeta"),
			})
			writeOwner(t, root, ownerDocumentVersion, "alpha", []Stream{platform})
			writeOwner(t, root, ownerDocumentVersion, "zeta", []Stream{tenant})

			if _, err := Load(root); err != nil {
				t.Fatalf("Load rejected tenant generation %d: %v", tenantGeneration, err)
			}
		})
	}
}

func TestLoadRejectsTenantGenerationWithoutPlatformFrontier(t *testing.T) {
	root := t.TempDir()
	tenant := validStream("iam", 10)
	tenant.Scope = "tenant"
	writeRoot(t, root, catalogVersion, []string{streamPath("iam")})
	writeOwner(t, root, ownerDocumentVersion, "iam", []Stream{tenant})

	_, err := Load(root)
	if err == nil || !strings.Contains(
		err.Error(),
		"tenant migration generation frontier 1 exceeds platform frontier 0",
	) {
		t.Fatalf("Load error = %v", err)
	}
}
