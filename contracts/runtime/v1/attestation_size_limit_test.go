package runtimev1

import "testing"

func TestAttestationDocumentLimitIsExact(t *testing.T) {
	if MaximumAttestationDocumentBytes != 64*1024 {
		t.Fatalf("attestation document limit = %d", MaximumAttestationDocumentBytes)
	}
}
