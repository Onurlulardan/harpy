package iam

import (
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestRBACRuntimeBuildsOnlyFromReleaseSealedCatalogs(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatalf("load catalogs: %v", err)
	}
	runtime, err := NewRBACRuntime(catalogs, time.Now)
	if err != nil || !runtime.Valid() {
		t.Fatalf("runtime = %+v, %v", runtime, err)
	}
}

func TestRBACRuntimeRejectsUnsealedCatalogOrMissingClock(t *testing.T) {
	if _, err := NewRBACRuntime(platformasync.WriteCatalogs{}, time.Now); err == nil {
		t.Fatal("unsealed catalogs were accepted")
	}
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatalf("load catalogs: %v", err)
	}
	if _, err := NewRBACRuntime(catalogs, nil); err == nil {
		t.Fatal("missing clock was accepted")
	}
}
