package postgres

import "testing"

func TestBillingSaaSBaseBindingIsAPIOnly(t *testing.T) {
	for _, profile := range []Profile{ProfileAPI, ProfileWorker} {
		pool := &Pool{profile: profile, contract: profileContract{runtime: true}}
		result, err := BindBillingSaaSBase(pool,
			func(BillingSaaSBaseDatabase) (bool, error) { return true, nil })
		if profile == ProfileAPI && (err != nil || !result) {
			t.Fatalf("API binding failed: %v", err)
		}
		if profile != ProfileAPI && err == nil {
			t.Fatal("worker obtained Billing SaaS base reader")
		}
	}
}
