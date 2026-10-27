package iam

import "testing"

func TestIAMRESTQueryRuntimeBuildsFinalOwnerSurface(t *testing.T) {
	runtime, err := NewIAMRESTQueryRuntime()
	if err != nil || !runtime.Valid() {
		t.Fatalf("runtime = %+v, %v", runtime, err)
	}
	if runtime.Users() == nil || runtime.Roles() == nil || runtime.Organizations() == nil {
		t.Fatal("catalog services are incomplete")
	}
	if runtime.Owner() != "iam" || runtime.Policy() == nil || runtime.Execution() == nil {
		t.Fatal("owner policy or execution surface is incomplete")
	}
}

func TestIAMRESTQueryRuntimeZeroValueFailsClosed(t *testing.T) {
	var runtime *IAMRESTQueryRuntime
	if runtime.Valid() || runtime.Users() != nil || runtime.Roles() != nil ||
		runtime.Organizations() != nil || runtime.Policy() != nil || runtime.Execution() != nil {
		t.Fatal("zero runtime exposed an owner surface")
	}
}
