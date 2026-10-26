package iam

import (
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestPermissionRequestRuntimeBuildsOwnedGraph(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewPermissionRequestRuntime(catalogs, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.Valid() {
		t.Fatal("permission request runtime is invalid")
	}
}
