package postgres

import (
	"errors"
	"testing"
)

func TestBillingLicenseAuthorityBinderAllowsOnlyWorkerAndLicenseOperator(t *testing.T) {
	for _, profile := range []Profile{
		ProfileWorker, ProfileOperatorLicenseReconcile,
	} {
		pool := &Pool{profile: profile, contract: profileContract{runtime: true}}
		value, err := BindBillingLicenseAuthority(pool,
			func(BillingLicenseAuthorityDatabase) (bool, error) {
				return true, nil
			})
		if err != nil || !value {
			t.Fatalf("profile %s rejected: %v", profile, err)
		}
	}
	pool := &Pool{profile: ProfileAPI, contract: profileContract{runtime: true}}
	if _, err := BindBillingLicenseAuthority(pool,
		func(BillingLicenseAuthorityDatabase) (bool, error) {
			return true, nil
		}); !errors.Is(err, ErrAPIStoreProfile) {
		t.Fatalf("API profile accepted: %v", err)
	}
}
