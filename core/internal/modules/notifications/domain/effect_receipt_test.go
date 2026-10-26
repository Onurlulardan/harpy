package domain

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"
)

func TestEffectReceiptReplaysOnlyExactCanonicalEffect(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	message, _ := NewV7(now, bytes.NewReader(bytes.Repeat([]byte{2}, 10)))
	recipient, _ := ParseID("11111111-1111-4111-8111-111111111111")
	candidate := EffectCandidate{Producer: "workflow", EffectKey: "instance/node/7",
		RecipientID: recipient, CanonicalDigest: sha256.Sum256([]byte("canonical-a"))}
	receipt, err := NewEffectReceipt(candidate, message, now, now.AddDate(1, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	got, outcome, err := receipt.Replay(candidate)
	if err != nil || got != message || outcome != EffectCreated {
		t.Fatalf("Replay() = %v, %s, %v", got, outcome, err)
	}
	candidate.CanonicalDigest = sha256.Sum256([]byte("canonical-b"))
	if _, _, err = receipt.Replay(candidate); err != ErrEffectReused {
		t.Fatalf("reuse error = %v", err)
	}
}
