package postgres

import "testing"

func TestTenancyCustomerWorkerBindingIsWorkerOnly(t *testing.T) {
	for _, profile := range []Profile{ProfileAPI, ProfileWorker} {
		pool := &Pool{profile: profile, contract: profileContract{runtime: true}}
		result, err := BindTenancyCustomerWorker(pool,
			func(TenancyCustomerWorkerDatabase) (bool, error) { return true, nil })
		if profile == ProfileWorker && (err != nil || !result) {
			t.Fatalf("worker binding failed: %v", err)
		}
		if profile != ProfileWorker && err == nil {
			t.Fatal("API obtained customer worker binding")
		}
	}
}
