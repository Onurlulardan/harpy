package v1adoption

import (
	"context"
	"testing"
)

func TestRetireRequiresZeroCompatibilityHits(t *testing.T) {
	target := &fixtureTarget{evidence: Evidence{Accepted: true,
		WriterFenced: true, CompatibilityHits: 1}}
	_, err := fixtureService(nil, target).Retire(context.Background(),
		fixtureTenantScope(), RetireRequest{RunID: fixtureID(1)})
	if err != ErrBlocked {
		t.Fatal(err)
	}
	target.evidence.CompatibilityHits = 0
	receipt, err := fixtureService(nil, target).Retire(context.Background(),
		fixtureTenantScope(), RetireRequest{RunID: fixtureID(1)})
	if err != nil || receipt.State != "RETIRED" {
		t.Fatal(err, receipt)
	}
}
