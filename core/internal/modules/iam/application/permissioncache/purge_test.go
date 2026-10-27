package permissioncache

import (
	"context"
	"testing"
)

type permissionCacheStoreStub struct{ result Result }

func (stub permissionCacheStoreStub) Purge(context.Context, Command) (Result, error) {
	return stub.result, nil
}

func TestPurgeRejectsInvalidCommandAndInvalidReceipt(t *testing.T) {
	application, _ := New(permissionCacheStoreStub{result: Result{AuthorizationClock: 2}})
	if _, err := application.Purge(t.Context(), Command{}); err != ErrInvalid {
		t.Fatalf("invalid command error = %v", err)
	}
	command := Command{TenantID: [16]byte{1}, ActorID: [16]byte{2}, OperationID: [16]byte{3},
		RequestSHA256: [32]byte{1}, TokenVersion: 1, AuthorizationVersion: 1,
		PlatformClock: 1}
	if _, err := application.Purge(t.Context(), command); err != ErrUnavailable {
		t.Fatalf("invalid receipt error = %v", err)
	}
}
