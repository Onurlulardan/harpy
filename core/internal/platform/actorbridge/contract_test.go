package actorbridge

import (
	"testing"
	"time"
)

func TestActorBridgeEvidenceTypesFailClosed(t *testing.T) {
	authority, err := NewAuthority(1, 2, 3)
	if err != nil || authority.IAMGroupGeneration() != 1 ||
		authority.IAMGeneration() != 2 || authority.TargetGeneration() != 3 {
		t.Fatal("valid actor bridge authority rejected")
	}
	if _, err := NewAuthority(0, 2, 3); err == nil {
		t.Fatal("invalid actor bridge authority accepted")
	}
	now := time.Unix(2_000_000_000, 0).UTC()
	receipt, err := NewReceipt(now)
	if err != nil || receipt.ConsumedAt() != now {
		t.Fatal("valid actor bridge receipt rejected")
	}
	if _, err := NewReceipt(time.Time{}); err == nil {
		t.Fatal("empty actor bridge receipt accepted")
	}
}

func TestActorBridgeReconciliationBindsCountToTimestamp(t *testing.T) {
	now := time.Unix(2_000_000_000, 0).UTC()
	result, err := NewReconciliation(2, &now)
	if err != nil || result.AccountedCount() != 2 {
		t.Fatal("valid reconciliation rejected")
	}
	latest, exists := result.LatestHitAt()
	if !exists || latest != now {
		t.Fatal("reconciliation timestamp differs")
	}
	if _, err := NewReconciliation(0, &now); err == nil {
		t.Fatal("zero-count reconciliation timestamp accepted")
	}
	if _, err := NewReconciliation(1, nil); err == nil {
		t.Fatal("positive reconciliation without timestamp accepted")
	}
}

func TestActorBridgeFaultClassificationIsTyped(t *testing.T) {
	err := NewFault(FaultReplay)
	if !FaultIs(err, FaultReplay) || FaultIs(err, FaultExpired) {
		t.Fatal("actor bridge fault classification differs")
	}
}
