package main

import "testing"

func TestReleaseIDsPreserveExistingEnvironment(t *testing.T) {
	const environment = "40000000-0000-4000-8000-000000000004"
	first := newReleaseIDsForEnvironment(environment)
	second := newReleaseIDsForEnvironment(environment)
	if first.environment != environment || second.environment != environment {
		t.Fatal("existing environment identity was not preserved")
	}
	if first.release == second.release || first.authorityManifest == second.authorityManifest {
		t.Fatal("new release reused release-scoped identities")
	}
}

func TestReleaseIDsCreateEnvironmentForFoundation(t *testing.T) {
	if ids := newReleaseIDsForEnvironment(""); ids.environment == "" {
		t.Fatal("foundation environment identity was not generated")
	}
}

func TestReleaseIDsKeepAppliedRelease(t *testing.T) {
	const applied = "01a183a1-8862-7315-aa0e-f8af7a1c08f1"
	ids := applyStoredReleaseID(newReleaseIDsForEnvironment(""), applied)
	if ids.release != applied {
		t.Fatalf("applied release id = %s", ids.release)
	}
}
