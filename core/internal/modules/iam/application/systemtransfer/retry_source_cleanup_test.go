package systemtransfer

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySourceCleanupPersistsBoundedBackoff(t *testing.T) {
	target, _, now := importCleanupFixture()
	store := &importCleanupStoreStub{}
	values := importCleanupValuesStub{now: now}
	_, err := RetrySourceCleanup(context.Background(), store, values,
		importCleanupJitterStub{value: 2}, target, errors.New("object unavailable"))
	if err != nil || store.retry.ErrorCode != SourceCleanupUnavailable ||
		store.retry.RetryAt != now.Add(2*time.Nanosecond) {
		t.Fatalf("RetrySourceCleanup() = %#v, %v", store.retry, err)
	}
	if got := sourceCleanupRetryLimit(1); got != 5*time.Second {
		t.Fatalf("first retry limit=%s", got)
	}
	if got := sourceCleanupRetryLimit(100); got != 15*time.Minute {
		t.Fatalf("maximum retry limit=%s", got)
	}
}

func TestRetrySourceCleanupTargetPinsCanonicalObjectKey(t *testing.T) {
	target, _, _ := importCleanupFixture()
	if target.Validate() != nil {
		t.Fatal("canonical target was rejected")
	}
	target.ObjectKey = "iam-system-import-sources/v1/other/1.fcart"
	if target.Validate() == nil {
		t.Fatal("cross-object cleanup target was accepted")
	}
}
