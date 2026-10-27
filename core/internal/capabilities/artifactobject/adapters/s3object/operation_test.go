package s3object

import (
	"context"
	"errors"
	"testing"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func TestPublicOperationsRejectInvalidRuntimeBoundaries(t *testing.T) {
	request, _ := uploadFixture()
	var missing *Store
	if _, err := missing.Upload(context.Background(), request); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("nil receiver error = %v", err)
	}
	store := testStore(t, &fakeAPI{})
	if _, err := store.Upload(nil, request); !errors.Is(err, ErrObjectInput) {
		t.Fatalf("nil context error = %v", err)
	}
	request.Body = nil
	if _, err := store.Upload(context.Background(), request); !errors.Is(err, ErrObjectInput) {
		t.Fatalf("nil body error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.DeleteExact(cancelled, artifactobject.Identity{
		Key: testObjectKey, Version: "version-1"}); !errors.Is(err, artifactobject.ErrUnavailable) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context error = %v", err)
	}
}
