package iam

import (
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestInvitationRuntimeBuildsOnlyWithReleaseCatalogsAndClock(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewInvitationRuntime(catalogs, func() time.Time {
		return time.Unix(2_000_000_000, 0).UTC()
	})
	if err != nil || !runtime.Valid() {
		t.Fatalf("runtime = %#v, %v", runtime, err)
	}
	if _, err := NewInvitationRuntime(catalogs, nil); err == nil {
		t.Fatal("nil invitation clock was accepted")
	}
}
