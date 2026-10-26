package postgres

import "testing"

func TestPlatformOpsRestoredLicenseBinderAllowsOnlyLicenseOperator(t *testing.T) {
	for _, profile := range []Profile{
		ProfileAPI, ProfileWorker, ProfileOperatorIdentityReconcile,
		ProfileOperatorLicenseReconcile,
	} {
		pool := &Pool{profile: profile, contract: profileContract{runtime: true}}
		_, err := BindPlatformOpsRestoredLicense(
			pool,
			func(PlatformOpsRestoredLicenseDatabase) (string, error) {
				return "bound", nil
			},
		)
		if (profile == ProfileOperatorLicenseReconcile) != (err == nil) {
			t.Fatalf("profile %q binding error=%v", profile, err)
		}
	}
}
