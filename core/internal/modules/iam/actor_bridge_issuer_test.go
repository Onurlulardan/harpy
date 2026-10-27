package iam

import (
	"context"
	"testing"

	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestActorBridgeIssuerIAMBuildsCurrentAuthorizationPipeline(t *testing.T) {
	runtime, err := NewActorBridgeIssuerIAM()
	if err != nil || !runtime.Valid() {
		t.Fatalf("runtime valid=%t err=%v", runtime.Valid(), err)
	}
	_, err = runtime.AuthorizeActorBridge(
		context.Background(), platformtx.TenantScope{}, "documents.items", "read", 1, 1,
	)
	if !platformactorbridge.FaultIs(err, platformactorbridge.FaultUnavailable) {
		t.Fatalf("unowned authorization error = %v", err)
	}
}
