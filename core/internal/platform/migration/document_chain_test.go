package migration

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

type documentChainCatalog struct {
	ContractVersion string               `json:"contract_version"`
	Owner           string               `json:"owner"`
	Scope           string               `json:"scope"`
	SourceDirectory string               `json:"source_directory"`
	Entries         []documentChainEntry `json:"entries"`
}

type documentChainEntry struct {
	OldPath           string `json:"old_path"`
	OldSHA256         string `json:"old_sha256"`
	ReplacementPath   string `json:"replacement_path"`
	ReplacementSHA256 string `json:"replacement_sha256"`
}

func TestNewSupersedingDocumentChainBindsThreeHistoryAliases(t *testing.T) {
	root := "core/internal/platform/example/migrations"
	base := "20260801010000_repair_example_v1.fresh.sql"
	contents := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four")}
	paths := []string{
		root + "/fresh/" + base,
		root + "/supersessions/fresh/" + base,
		root + "/supersessions/v2/fresh/" + base,
		root + "/supersessions/v3/fresh/" + base,
	}
	document := documentChainCatalog{
		ContractVersion: migrationsupersession.ContractVersionV3,
		Owner:           "platform_example", Scope: "platform", SourceDirectory: root,
	}
	for level := 0; level < 3; level++ {
		document.Entries = append(document.Entries, documentChainEntry{
			OldPath: paths[level], OldSHA256: digestDocumentChain(contents[level]),
			ReplacementPath:   paths[level+1],
			ReplacementSHA256: digestDocumentChain(contents[level+1]),
		})
	}
	addAdoptedDocumentChain(&document, root, contents)
	sort.Slice(document.Entries, func(left, right int) bool {
		return document.Entries[left].OldPath < document.Entries[right].OldPath
	})
	encoded, err := attestation.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := migrationsupersession.DecodeCatalog(append(encoded, '\n'))
	if err != nil {
		t.Fatal(err)
	}
	value, err := NewSupersedingDocumentChain(
		"platform_example", "platform", paths[3], contents[3], catalog.Lineage(paths[3]),
	)
	if err != nil || len(value.superseded) != 3 {
		t.Fatalf("three-alias document binding differs: %d %v", len(value.superseded), err)
	}
}

func addAdoptedDocumentChain(
	document *documentChainCatalog, root string, contents [][]byte,
) {
	base := "20260801010000_repair_example_v1.v1_adopted.sql"
	paths := []string{
		root + "/v1_adopted/" + base,
		root + "/supersessions/v1_adopted/" + base,
		root + "/supersessions/v2/v1_adopted/" + base,
		root + "/supersessions/v3/v1_adopted/" + base,
	}
	for level := 0; level < 3; level++ {
		document.Entries = append(document.Entries, documentChainEntry{
			OldPath: paths[level], OldSHA256: digestDocumentChain(contents[level]),
			ReplacementPath:   paths[level+1],
			ReplacementSHA256: digestDocumentChain(contents[level+1]),
		})
	}
}

func digestDocumentChain(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}
