package iam

import "testing"

func TestPermissionCacheRuntimeRejectsMissingDependencies(t *testing.T) {
	if _, err := NewPermissionCacheRuntime(nil, nil); err == nil {
		t.Fatal("missing dependencies were accepted")
	}
}
