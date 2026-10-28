package migrationstreamcatalog

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestRepositoryStreamsEqualLegacyFlatContract(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	streams := manifest.Streams()
	legacy := legacyFlatStreams()
	if len(streams) < len(legacy) || !reflect.DeepEqual(streams[:len(legacy)], legacy) {
		t.Fatalf("v2 stream prefix differs from the frozen legacy flat contract: %#v", streams)
	}
	for _, stream := range streams[len(legacy):] {
		if stream.VersionFrom <= legacy[len(legacy)-1].VersionTo {
			t.Fatalf("post-legacy stream is not append-only: %#v", stream)
		}
	}
}

func legacyFlatStreams() []Stream {
	return []Stream{
		legacyStream("platform_migrator", "core/internal/platform/migrator/migrations", "platform", 20260713010000, 20260713015200, "foundation", 1),
		legacyStream("platform_cutover", "core/internal/platform/cutover/migrations", "platform", 20260713020000, 20260713023300, "foundation", 1),
		legacyStream("tenancy", "core/internal/modules/tenancy/migrations/platform", "platform", 20260713030000, 20260713031600, "foundation", 1),
		legacyStream("platform_migrator", "core/internal/platform/migrator/migrations", "platform", 20260713040000, 20260713042300, "foundation", 1),
		legacyStream("platform_migrator", "core/internal/platform/migrator/migrations", "platform", 20260714005900, 20260714010300, "foundation", 1),
		legacyStream("platform_migrator", "core/internal/platform/migrator/migrations", "platform", 20260714010400, 20260714010400, "expand", 2),
		legacyStream("platform_migrator", "core/internal/platform/migrator/migrations", "platform", 20260714010500, 20260714010900, "expand", 2),
		legacyStream("tenancy", "core/internal/modules/tenancy/migrations/tenant", "tenant", 20260714011000, 20260714011200, "expand", 2),
		legacyStream("platform_cutover", "core/internal/platform/cutover/migrations", "platform", 20260714012000, 20260714012900, "expand", 2),
		legacyStream("platform_migrator", "core/internal/platform/migrator/migrations", "platform", 20260714012910, 20260714012930, "expand", 2),
		legacyStream("platform_cutover", "core/internal/platform/cutover/migrations", "platform", 20260714013400, 20260714014100, "expand", 2),
		legacyStream("platform_cutover", "core/internal/platform/cutover/migrations", "platform", 20260715010000, 20260715010600, "expand", 2),
		legacyStream("platform_async", "core/internal/platform/async/migrations", "platform", 20260715020000, 20260715021600, "expand", 3),
		legacyStream("platform_cutover", "core/internal/platform/cutover/migrations", "platform", 20260715021605, 20260715021605, "expand", 3),
		legacyStream("platform_async", "core/internal/platform/async/migrations", "platform", 20260715021610, 20260715022010, "expand", 3),
		legacyStream("platform_cutover", "core/internal/platform/cutover/migrations", "platform", 20260716005800, 20260716005855, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/platform", "platform", 20260716010000, 20260716010000, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/tenant", "tenant", 20260716011000, 20260716011900, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/platform", "platform", 20260716012000, 20260716012010, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/platform", "platform", 20260716012020, 20260716012059, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/tenant", "tenant", 20260716012100, 20260716012900, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/tenant", "tenant", 20260716012920, 20260716012926, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/tenant", "tenant", 20260716012950, 20260716012956, "expand", 4),
		legacyStream("iam", "core/internal/modules/iam/migrations/tenant", "tenant", 20260716013020, 20260716013150, "expand", 4),
	}
}

func legacyStream(
	owner string,
	directory string,
	scope string,
	versionFrom int64,
	versionTo int64,
	phase string,
	databaseGeneration int64,
) Stream {
	return Stream{
		Owner: owner, Directory: directory, Scope: scope,
		VersionFrom: versionFrom, VersionTo: versionTo, Phase: phase,
		TargetDatabaseGeneration:     databaseGeneration,
		NextMinimumRuntimeGeneration: 1,
	}
}
