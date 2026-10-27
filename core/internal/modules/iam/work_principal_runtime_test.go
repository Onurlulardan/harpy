package iam

import "testing"

func TestWorkPrincipalRuntimeBuildsIAMOwnedReader(t *testing.T) {
	runtime, err := NewWorkPrincipalRuntime()
	if err != nil {
		t.Fatal(err)
	}
	if runtime == nil || runtime.reader == nil {
		t.Fatal("Work principal runtime is unavailable")
	}
}
