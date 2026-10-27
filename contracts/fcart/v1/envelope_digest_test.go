package fcartv1

import "testing"

func TestKeyEnvelopeDigestIsCanonicalAndOwned(t *testing.T) {
	envelope := KeyEnvelope{KEKID: "artifact-2026-07", Nonce: [12]byte{1},
		Ciphertext: make([]byte, 48)}
	envelope.Ciphertext[0] = 1
	digest, err := envelope.SHA256()
	if err != nil || digest == [32]byte{} {
		t.Fatalf("SHA256() = %x, %v", digest, err)
	}
	envelope.Ciphertext[0] = 2
	changed, err := envelope.SHA256()
	if err != nil || changed == digest {
		t.Fatal("envelope digest did not bind ciphertext")
	}
}
