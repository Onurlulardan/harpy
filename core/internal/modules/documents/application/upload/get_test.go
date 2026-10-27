package upload

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestGetBoundReturnsRedactedPersistedState(t *testing.T) {
	store := uploadStore{get: func(_ context.Context, id domain.ID) (Snapshot, error) {
		if id != uploadHandle {
			t.Fatal("unexpected upload id")
		}
		return uploadSnapshot(), nil
	}}
	result, err := getBound(context.Background(), store, uploadHandle)
	if err != nil || result.ID != uploadHandle || result.State != domain.UploadCreated {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
