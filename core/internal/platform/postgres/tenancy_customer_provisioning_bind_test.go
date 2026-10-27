package postgres

import (
	"errors"
	"testing"
)

func TestBindTenancyCustomerProvisioningAllowsRuntimeProfiles(t *testing.T) {
	t.Parallel()
	for _, profile := range []Profile{ProfileAPI, ProfileWorker} {
		pool := &Pool{profile: profile, contract: profileContract{runtime: true}}
		bound, err := BindTenancyCustomerProvisioning(pool,
			func(TenancyCustomerProvisioningDatabase) (bool, error) { return true, nil })
		if err != nil || !bound {
			t.Fatalf("profile %s bound=%v err=%v", profile, bound, err)
		}
	}
}

func TestBindTenancyCustomerProvisioningRejectsOtherProfiles(t *testing.T) {
	t.Parallel()
	pool := &Pool{profile: ProfileOperatorInitialize, contract: profileContract{runtime: true}}
	_, err := BindTenancyCustomerProvisioning(pool,
		func(TenancyCustomerProvisioningDatabase) (bool, error) { return true, nil })
	if !errors.Is(err, ErrTenancyPolicyProfile) {
		t.Fatalf("error = %v", err)
	}
}
