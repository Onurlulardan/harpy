package main

import (
	"strings"
	"testing"
)

func TestPlaceholderSeedNeedsLiveBinding(t *testing.T) {
	row := baselineIdentityRow{
		environmentID: baselineSeedEnvironmentID, releaseID: baselineSeedReleaseID,
		origin: "fresh", authority: baselineSeedAuthority,
		systemID: baselineSeedSystemID, databaseOID: baselineSeedDatabaseOID,
	}
	live := liveDatabaseIdentity{systemID: "1", databaseOID: 1}
	if !placeholderSeedNeedsBinding(row, live) {
		t.Fatal("dumped placeholder was left unbound")
	}
	live.systemID = baselineSeedSystemID
	live.databaseOID = uint32(baselineSeedDatabaseOID)
	if placeholderSeedNeedsBinding(row, live) {
		t.Fatal("placeholder already bound to this database was selected again")
	}
	row.systemID = "2"
	if placeholderSeedNeedsBinding(row, liveDatabaseIdentity{systemID: "1", databaseOID: 1}) {
		t.Fatal("a live identity was treated as the dumped placeholder")
	}
}

func TestBindBaselineSeedStatementTargetsPlaceholder(t *testing.T) {
	if !strings.Contains(bindBaselineSeedSQL, "SET database_system_identifier=$1, database_oid=$2") ||
		!strings.Contains(bindBaselineSeedSQL, "database_system_identifier=$6") ||
		!strings.Contains(bindBaselineSeedSQL, "database_oid=$7") ||
		strings.Contains(bindBaselineSeedSQL, "SET release_id") ||
		strings.Contains(bindBaselineSeedSQL, "SET environment_id") {
		t.Fatal("rebind must change only the dumped physical identity")
	}
}

func TestEnvironmentIdentityTriggerIsRestored(t *testing.T) {
	if assumeEnvironmentOwnerSQL != "SET LOCAL ROLE harpy_owner" ||
		!strings.Contains(suspendEnvironmentTriggerSQL, "DISABLE TRIGGER environment_identity_reject_mutation") ||
		!strings.Contains(resumeEnvironmentTriggerSQL, "ENABLE TRIGGER environment_identity_reject_mutation") {
		t.Fatal("placeholder rebind must suspend and restore the same immutable trigger")
	}
}

func TestUnboundBaselineSeedMatchesFreshDumpRow(t *testing.T) {
	if !unboundBaselineSeed(
		baselineSeedEnvironmentID, baselineSeedReleaseID, "fresh", baselineSeedAuthority,
	) {
		t.Fatal("fresh baseline seed was not recognized")
	}
	if unboundBaselineSeed(
		baselineSeedEnvironmentID, baselineSeedReleaseID, "fresh", baselineSeedSystemID,
	) {
		t.Fatal("unrelated authority was treated as the baseline seed")
	}
	if unboundBaselineSeed(
		"018f0000-0000-7000-8000-000000000103", baselineSeedReleaseID, "fresh",
		baselineSeedAuthority,
	) {
		t.Fatal("another environment was treated as the baseline seed")
	}
	if baselineSeedSystemID != "7703580167595228103" || baselineSeedDatabaseOID != 51019 {
		t.Fatal("baseline seed physical identity changed")
	}
}
