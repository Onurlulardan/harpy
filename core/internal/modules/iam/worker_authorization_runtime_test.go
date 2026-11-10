package iam

import "testing"

func TestWorkerAuthorizationRuntimeRequiresCanonicalEvidence(t *testing.T) {
	if _, err := newWorkerAuthorizationRuntime(nil); err == nil {
		t.Fatal("nil IAM evidence was accepted")
	}
}
