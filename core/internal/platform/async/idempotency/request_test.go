package idempotency

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRequestConstructorsSealRetentionInputs(t *testing.T) {
	t.Parallel()
	digest := strings.Repeat("a", 64)
	write, err := NewWriteRequest("key", digest, 9, InteractiveWrite)
	if err != nil || write.AsyncJobID() != nil || write.RequestedRetainUntil() != nil {
		t.Fatalf("write request = %#v, %v", write, err)
	}
	retainUntil := time.Now().UTC().Add(40 * 24 * time.Hour)
	jobID := uuid.New()
	handle, err := NewAsyncHandleRequest("key", digest, 9, jobID, retainUntil)
	if err != nil || handle.RetentionClass() != AsyncHandle ||
		handle.AsyncJobID() == nil || *handle.AsyncJobID() != jobID ||
		handle.RequestedRetainUntil() == nil {
		t.Fatalf("async request = %#v, %v", handle, err)
	}
}

func TestRequestRejectsCallerSelectedOrInvalidRetention(t *testing.T) {
	t.Parallel()
	digest := strings.Repeat("a", 64)
	if _, err := NewWriteRequest("key", digest, 1, AsyncHandle); err == nil {
		t.Fatal("async class accepted without job and horizon")
	}
	if _, err := NewAsyncHandleRequest("key", digest, 1, uuid.Nil, time.Now()); err == nil {
		t.Fatal("nil job accepted")
	}
	if _, err := NewAsyncHandleRequest("key", digest, 1, uuid.New(), time.Time{}); err == nil {
		t.Fatal("zero artifact horizon accepted")
	}
	if _, err := NewWriteRequest("", digest, 1, InteractiveWrite); err == nil {
		t.Fatal("empty key accepted")
	}
	if _, err := NewWriteRequest("key", strings.Repeat("A", 64), 1, InteractiveWrite); err == nil {
		t.Fatal("non-canonical request digest accepted")
	}
	if _, err := NewWriteRequest("key", digest, 0, InteractiveWrite); err == nil {
		t.Fatal("zero authorization version accepted")
	}
}
