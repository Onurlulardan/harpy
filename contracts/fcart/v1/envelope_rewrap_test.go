package fcartv1

import (
	"bytes"
	"context"
	"testing"
)

func TestEnvelopeKeyringRewrapsRetainedKeyToCurrent(t *testing.T) {
	oldKey, _ := NewKEK("artifact-2026-01", bytes.Repeat([]byte{1}, DEKBytes))
	current, _ := NewKEK("artifact-2026-07", bytes.Repeat([]byte{2}, DEKBytes))
	identity := envelopeIdentity()
	oldRing, _ := NewEnvelopeKeyring(oldKey, nil,
		bytes.NewReader(bytes.Repeat([]byte{3}, 256)))
	dek, oldEnvelope, err := oldRing.Generate(context.Background(), identity)
	if err != nil {
		t.Fatal(err)
	}
	ring, _ := NewEnvelopeKeyring(current, []KEK{oldKey},
		bytes.NewReader(bytes.Repeat([]byte{4}, 256)))
	if !ring.SupportsKEK(oldKey.ID()) || !ring.SupportsKEK(current.ID()) ||
		ring.SupportsKEK("missing") {
		t.Fatal("keyring support inventory differs")
	}
	rewrapped, changed, err := ring.Rewrap(context.Background(), identity, oldEnvelope)
	if err != nil || !changed || rewrapped.KEKID != current.ID() ||
		bytes.Equal(rewrapped.Ciphertext, oldEnvelope.Ciphertext) {
		t.Fatalf("Rewrap() = %#v, %t, %v", rewrapped, changed, err)
	}
	opened, err := ring.Unwrap(context.Background(), identity, rewrapped)
	if err != nil || opened != dek {
		t.Fatalf("Unwrap(rewrapped) = %v, %v", opened, err)
	}
}

func TestEnvelopeKeyringVerifiesCurrentEnvelopeBeforeNoop(t *testing.T) {
	current, _ := NewKEK("artifact-2026-07", bytes.Repeat([]byte{2}, DEKBytes))
	identity := envelopeIdentity()
	ring, _ := NewEnvelopeKeyring(current, nil,
		bytes.NewReader(bytes.Repeat([]byte{4}, 256)))
	_, envelope, _ := ring.Generate(context.Background(), identity)
	envelope.Ciphertext[0] ^= 1
	if _, _, err := ring.Rewrap(context.Background(), identity, envelope); err == nil {
		t.Fatal("Rewrap accepted a forged current envelope")
	}
}
