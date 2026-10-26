package iam

import "testing"

func TestIdentityRuntimeBuildsCanonicalApplicationGraph(t *testing.T) {
	runtime, err := NewIdentityRuntime()
	if err != nil || !runtime.Valid() {
		t.Fatalf("runtime = %+v, %v", runtime, err)
	}
}
