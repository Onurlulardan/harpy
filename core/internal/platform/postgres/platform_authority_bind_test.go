package postgres

import "testing"

func TestPlatformAuthorityStoreAllowsAuthorizedRuntimeProfiles(t *testing.T) {
	for _, profile := range []Profile{ProfileAPI, ProfileWorker, ProfileOperatorInitialize} {
		pool := &Pool{profile: profile, contract: profileContract{runtime: true}}
		result, err := BindPlatformAuthorityStore(pool,
			func(PlatformAuthorityTransactions) (bool, error) { return true, nil })
		if err != nil || !result {
			t.Fatalf("%s binding failed: %v", profile, err)
		}
	}
}

func TestPlatformAuthorityStoreRejectsOtherProfiles(t *testing.T) {
	pool := &Pool{profile: ProfileMigratorDeploy,
		contract: profileContract{runtime: true}}
	_, err := BindPlatformAuthorityStore(pool,
		func(PlatformAuthorityTransactions) (bool, error) { return true, nil })
	if err == nil {
		t.Fatal("migrator profile obtained platform authority store")
	}
}
