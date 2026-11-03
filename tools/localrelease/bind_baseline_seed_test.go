package main

import "testing"

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
