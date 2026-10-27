package migration

import (
	"strings"
	"testing"
)

func TestBindHistorySupersessionsSelectsOnlyDatabaseOrigin(t *testing.T) {
	_, selected, documents := testMigrationFixture(t)
	documents[0].superseded = []documentIdentity{{
		path:   "legacy/migrations/fresh/20260713010000_create_foundation.fresh.sql",
		sha256: strings.Repeat("a", 64),
	}}
	documents[1].superseded = []documentIdentity{{
		path: "legacy/migrations/v1_adopted/" +
			"20260713010000_create_foundation.v1_adopted.sql",
		sha256: strings.Repeat("b", 64),
	}}
	byScope, err := bindHistorySupersessions(selected, documents)
	if err != nil {
		t.Fatal(err)
	}
	values := byScope["platform"]
	if len(values) != 1 || values[0].Version() != testVersion ||
		values[0].CurrentPath() != testFreshPath ||
		values[0].SupersededPath() != documents[0].superseded[0].path ||
		values[0].SupersededSHA256() != documents[0].superseded[0].sha256 {
		t.Fatalf("history supersession binding differs: %#v", values)
	}
}

func TestBindHistorySupersessionsRejectsDuplicateVersion(t *testing.T) {
	_, selected, documents := testMigrationFixture(t)
	documents[0].superseded = []documentIdentity{{
		path: "legacy/first.sql", sha256: strings.Repeat("a", 64),
	}}
	duplicate := documents[0]
	duplicate.superseded = []documentIdentity{{
		path: "legacy/second.sql", sha256: strings.Repeat("b", 64),
	}}
	if _, err := bindHistorySupersessions(
		selected, append(documents, duplicate),
	); err == nil {
		t.Fatal("duplicate history supersession version was accepted")
	}
}

func TestBindHistorySupersessionsEmitsEveryChainAlias(t *testing.T) {
	_, selected, documents := testMigrationFixture(t)
	documents[0].superseded = []documentIdentity{
		{path: "legacy/first.sql", sha256: strings.Repeat("a", 64)},
		{path: "legacy/second.sql", sha256: strings.Repeat("b", 64)},
		{path: "legacy/third.sql", sha256: strings.Repeat("c", 64)},
	}
	byScope, err := bindHistorySupersessions(selected, documents)
	if err != nil {
		t.Fatal(err)
	}
	values := byScope["platform"]
	if len(values) != 3 ||
		values[0].SupersededPath() != "legacy/first.sql" ||
		values[1].SupersededPath() != "legacy/second.sql" ||
		values[2].SupersededPath() != "legacy/third.sql" {
		t.Fatalf("history chain aliases differ: %#v", values)
	}
}
