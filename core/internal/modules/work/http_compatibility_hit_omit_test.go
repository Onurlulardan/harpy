package work

import "testing"

func TestOmitCompatibilityHitSkipsBrowserJWTClients(t *testing.T) {
	if !omitCompatibilityHit(HTTPCompatibilityHit{}) {
		t.Fatal("empty cohort headers must skip compatibility hit persistence")
	}
	if omitCompatibilityHit(HTTPCompatibilityHit{ConsumerID: "frontend-v3", BuildID: "dev"}) {
		t.Fatal("signed cohort headers must record compatibility hits")
	}
}
