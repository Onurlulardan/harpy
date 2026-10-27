package usage

import "testing"

func TestResolveExpiryEvidenceMatrix(t *testing.T) {
	tests := []struct {
		started  bool
		evidence EffectEvidence
		want     ReservationState
		wantErr  bool
	}{
		{false, EffectUnknown, Released, false},
		{false, EffectConfirmed, "", true},
		{false, NoEffectConfirmed, Released, false},
		{true, EffectUnknown, ReconcileRequired, false},
		{true, EffectConfirmed, Committed, false},
		{true, NoEffectConfirmed, Released, false},
	}
	for _, test := range tests {
		got, err := ResolveExpiry(test.started, test.evidence)
		if (err != nil) != test.wantErr || got != test.want {
			t.Fatalf("started=%v evidence=%s: got %s, %v", test.started, test.evidence, got, err)
		}
	}
}

func TestResolveExpiryRejectsUnknownEvidence(t *testing.T) {
	if _, err := ResolveExpiry(true, "invalid"); err != ErrExpiry {
		t.Fatalf("expected ErrExpiry, got %v", err)
	}
}
