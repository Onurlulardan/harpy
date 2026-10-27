package postgres

import (
	"errors"
	"testing"
)

func TestBillingLicenseAPIBinderAllowsOnlyAPIProfile(t *testing.T) {
	apiPool := &Pool{
		profile:  ProfileAPI,
		contract: profileContract{runtime: true},
	}
	value, err := BindBillingLicenseAPI(
		apiPool,
		func(BillingLicenseAPIDatabase) (bool, error) {
			return true, nil
		},
	)
	if err != nil || !value {
		t.Fatalf("API profile rejected: value=%t err=%v", value, err)
	}
	for _, profile := range []Profile{
		ProfileWorker, ProfileOperatorLicenseReconcile,
	} {
		pool := &Pool{
			profile:  profile,
			contract: profileContract{runtime: true},
		}
		_, err := BindBillingLicenseAPI(
			pool,
			func(BillingLicenseAPIDatabase) (bool, error) {
				return true, nil
			},
		)
		if !errors.Is(err, ErrAPIStoreProfile) {
			t.Fatalf("profile %s accepted: %v", profile, err)
		}
	}
}
