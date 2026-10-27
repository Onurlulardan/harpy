package v1adoption

import (
	"context"
	"testing"
)

func TestReopenRequiresAcceptedFencedEvidence(t *testing.T) {
	target := &fixtureTarget{evidence: Evidence{Accepted: true, WriterFenced: true,
		SourceSHA256: [32]byte{1}, TargetSHA256: [32]byte{2}}}
	request := ReverseRequest{RunID: fixtureID(1), ReconciliationID: fixtureID(12),
		ApprovedBy: fixtureID(13), SourceGeneration: 10, TargetGeneration: 11}
	receipt, err := fixtureService(nil, target).Reopen(
		context.Background(), fixtureTenantScope(), request)
	if err != nil || receipt.State != "REVERSED" {
		t.Fatal(err, receipt)
	}
}
