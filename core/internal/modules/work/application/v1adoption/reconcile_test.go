package v1adoption

import (
	"context"
	"testing"
)

func TestReconcileRequiresExactTransformedEvidence(t *testing.T) {
	target := &fixtureTarget{}
	evidence, err := fixtureService(nil, target).Reconcile(
		context.Background(), fixtureTenantScope(), fixtureRequest())
	if err != nil || !evidence.Accepted || evidence.TargetSHA256 != targetDigest(nil) {
		t.Fatal(err, evidence)
	}
}
