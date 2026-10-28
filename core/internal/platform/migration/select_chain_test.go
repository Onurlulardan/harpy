package migration

import (
	"strings"
	"testing"
)

func TestSelectFreshPreservesSupersessionChain(t *testing.T) {
	const version = "20260730120104_repair_installation_layout_publications_v1.fresh.sql"
	root := "core/internal/modules/system_resource_layout/migrations/tenant/"
	base := freshTestDocument(t, root+"fresh/"+version, 1)
	terminal := freshTestDocument(t, root+"supersessions/v3/fresh/"+version, 2)
	terminal.superseded = []documentIdentity{{
		path: root + "fresh/" + version, sha256: strings.Repeat("a", 64),
	}}
	selected, err := SelectFreshDocuments([]Document{base, terminal})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || len(selected[0].superseded) != 1 ||
		selected[0].superseded[0].path != terminal.superseded[0].path {
		t.Fatalf("supersession chain = %#v", selected)
	}
}
