package migrationstreamcatalog

import (
	"path/filepath"
	"testing"
)

func TestRepositoryTenantFrontierAdvancesHaveSequentialPlatformFences(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	streams := manifest.Streams()
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner: "iam", Directory: "core/internal/modules/iam/migrations/tenant",
		Scope: "tenant", VersionFrom: 20260728104000, VersionTo: 20260728104000,
		Phase: "expand", TargetDatabaseGeneration: 67,
		NextMinimumRuntimeGeneration: 23,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "platform_cutover",
		Directory: "core/internal/platform/cutover/migrations",
		Scope:     "platform", VersionFrom: 20260728104010, VersionTo: 20260728104010,
		Phase: "expand", TargetDatabaseGeneration: 67,
		NextMinimumRuntimeGeneration: 23,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner: "catalog", Directory: "core/internal/modules/catalog/migrations/tenant",
		Scope: "tenant", VersionFrom: 20260728105100, VersionTo: 20260728105100,
		Phase: "expand", TargetDatabaseGeneration: 68,
		NextMinimumRuntimeGeneration: 23,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "platform_cutover",
		Directory: "core/internal/platform/cutover/migrations",
		Scope:     "platform", VersionFrom: 20260728105110, VersionTo: 20260728105110,
		Phase: "expand", TargetDatabaseGeneration: 68,
		NextMinimumRuntimeGeneration: 23,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner: "iam", Directory: "core/internal/modules/iam/migrations/tenant",
		Scope: "tenant", VersionFrom: 20260730113000, VersionTo: 20260730113000,
		Phase: "expand", TargetDatabaseGeneration: 73,
		NextMinimumRuntimeGeneration: 24,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "platform_migrator",
		Directory: "core/internal/platform/migrator/migrations",
		Scope:     "platform", VersionFrom: 20260730113100, VersionTo: 20260730113100,
		Phase: "expand", TargetDatabaseGeneration: 74,
		NextMinimumRuntimeGeneration: 24,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner: "iam", Directory: "core/internal/modules/iam/migrations/tenant",
		Scope: "tenant", VersionFrom: 20260730120000, VersionTo: 20260730120020,
		Phase: "expand", TargetDatabaseGeneration: 79,
		NextMinimumRuntimeGeneration: 24,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "system_resource_layout",
		Directory: "core/internal/modules/system_resource_layout/migrations/tenant",
		Scope:     "tenant", VersionFrom: 20260730120100, VersionTo: 20260730120104,
		Phase: "expand", TargetDatabaseGeneration: 79,
		NextMinimumRuntimeGeneration: 24,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "platform_migrator",
		Directory: "core/internal/platform/migrator/migrations",
		Scope:     "platform", VersionFrom: 20260730120200, VersionTo: 20260730120200,
		Phase: "expand", TargetDatabaseGeneration: 80,
		NextMinimumRuntimeGeneration: 24,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "platform_migrator",
		Directory: "core/internal/platform/migrator/migrations",
		Scope:     "platform", VersionFrom: 20260730120300, VersionTo: 20260730120304,
		Phase: "expand", TargetDatabaseGeneration: 81,
		NextMinimumRuntimeGeneration: 24,
	})
	assertRepositoryFrontierStream(t, streams, Stream{
		Owner:     "platform_migrator",
		Directory: "core/internal/platform/migrator/migrations",
		Scope:     "platform", VersionFrom: 20260730120305, VersionTo: 20260730120314,
		Phase: "expand", TargetDatabaseGeneration: 82,
		NextMinimumRuntimeGeneration: 24,
	})
}

func assertRepositoryFrontierStream(
	t *testing.T,
	streams []Stream,
	want Stream,
) {
	t.Helper()
	for _, stream := range streams {
		if stream.VersionFrom == want.VersionFrom {
			if stream != want {
				t.Fatalf("migration stream %d = %#v, want %#v",
					want.VersionFrom, stream, want)
			}
			return
		}
	}
	t.Fatalf("migration stream %d is absent", want.VersionFrom)
}
