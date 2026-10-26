package migrationsupersession

import (
	"sort"
	"strings"
	"testing"
)

func TestDecodeV3AcceptsContiguousMixedDepthChains(t *testing.T) {
	encoded, _ := v3ChainFixture(t)
	var decoded catalogDocument
	decodeCatalogDocument(t, encoded, &decoded)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		base := "20260721010001_compile_other_v1." + origin + ".sql"
		decoded.Entries = append(decoded.Entries, entryDocument{
			OldPath:           fixtureDirectory + "/" + origin + "/" + base,
			OldSHA256:         strings.Repeat("a", 64),
			ReplacementPath:   fixtureDirectory + "/supersessions/" + origin + "/" + base,
			ReplacementSHA256: strings.Repeat("b", 64),
		})
	}
	sort.Slice(decoded.Entries, func(left, right int) bool {
		return decoded.Entries[left].OldPath < decoded.Entries[right].OldPath
	})
	if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err != nil {
		t.Fatalf("contiguous mixed-depth catalog failed: %v", err)
	}
}

func TestDecodeV3RejectsThirdLinkDriftAndUnorderedEntries(t *testing.T) {
	encoded, _ := v3ChainFixture(t)
	for name, mutate := range map[string]func(*catalogDocument){
		"third link hash": func(document *catalogDocument) {
			for index := range document.Entries {
				if strings.Contains(document.Entries[index].ReplacementPath, "/v3/fresh/") {
					document.Entries[index].OldSHA256 = strings.Repeat("f", 64)
				}
			}
		},
		"origin terminal hash": func(document *catalogDocument) {
			for index := range document.Entries {
				if strings.Contains(document.Entries[index].ReplacementPath, "/v3/v1_adopted/") {
					document.Entries[index].ReplacementSHA256 = strings.Repeat("e", 64)
				}
			}
		},
		"unordered entries": func(document *catalogDocument) {
			document.Entries[0], document.Entries[1] = document.Entries[1], document.Entries[0]
		},
		"non-exact v3 path": func(document *catalogDocument) {
			for index := range document.Entries {
				if strings.Contains(document.Entries[index].ReplacementPath, "/v3/fresh/") {
					document.Entries[index].ReplacementPath = strings.Replace(
						document.Entries[index].ReplacementPath, "/v3/", "/v4/", 1,
					)
				}
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			var decoded catalogDocument
			decodeCatalogDocument(t, encoded, &decoded)
			mutate(&decoded)
			if _, err := DecodeCatalog(marshalCatalogDocument(t, decoded)); err == nil {
				t.Fatal("invalid v3 catalog was accepted")
			}
		})
	}
}

func TestSelectV3RejectsMissingAndOrphanTerminalSources(t *testing.T) {
	encoded, sources := v3ChainFixture(t)
	catalog, err := DecodeCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	missing := append([]Source(nil), sources[:3]...)
	missing = append(missing, sources[4:]...)
	if _, err = Select(catalog, missing); err == nil {
		t.Fatal("missing v3 terminal source was accepted")
	}
	orphan := append([]Source(nil), sources...)
	orphan = append(orphan, Source{
		Path: fixtureDirectory + "/supersessions/v3/fresh/" +
			"20260721010001_orphan_v1.fresh.sql",
		Content: []byte("-- +goose Up\nSELECT 2;\n"),
	})
	if _, err = Select(catalog, orphan); err == nil {
		t.Fatal("orphan v3 terminal source was accepted")
	}
}
