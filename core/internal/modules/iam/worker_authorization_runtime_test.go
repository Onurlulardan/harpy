package iam

import (
	"context"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestWorkerAuthorizationRuntimeRequiresCanonicalEvidence(t *testing.T) {
	if _, err := newWorkerAuthorizationRuntime(nil); err == nil {
		t.Fatal("nil IAM evidence was accepted")
	}
}

func TestAnalyticsWorkerAuthorizationRuntimeExposesPinnedActorGate(t *testing.T) {
	runtime, err := NewAnalyticsWorkerAuthorizationRuntime()
	if err != nil || runtime == nil {
		t.Fatalf("build Analytics worker IAM runtime: %v", err)
	}
	allowed, permissionID, tokenVersion, authorizationVersion, clock, err :=
		runtime.AuthorizePinned(context.Background(), platformtx.TenantScope{},
			"analytics.dashboard", "export", 7)
	if err == nil || allowed || permissionID != [16]byte{} || tokenVersion != 0 ||
		authorizationVersion != 0 || clock != 0 {
		t.Fatalf("unowned pinned gate = %t/%s/%d/%d/%d err=%v",
			allowed, permissionID, tokenVersion, authorizationVersion, clock, err)
	}
}
