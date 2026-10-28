package main

import "testing"

func TestEmptyEnvironmentInstallsCurrentRelease(t *testing.T) {
	err := verifyNewReleaseStage(
		t.Context(), nil, currentReleaseStatusTestContract(), "",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFoundationBaselineStillRejectsReplacement(t *testing.T) {
	contract := currentReleaseStatusTestContract()
	contract.MigrationManifestSHA256 = foundationMigrationDigest
	err := verifyNewReleaseStage(t.Context(), nil, contract, "018f0000-0000-7000-8000-000000000001")
	if err == nil {
		t.Fatal("existing baseline was replaceable")
	}
}
