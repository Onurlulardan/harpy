package application

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func TestDeleteRequiresExpectedDraftRevision(t *testing.T) {
	store := &fixtureStore{}
	store.delete = func(_ context.Context, id uuid.UUID, expected int64) error {
		if id != fixtureRevisionID || expected != 3 {
			t.Fatal("delete lost optimistic revision")
		}
		return nil
	}
	err := fixtureApplication(t, store).Delete(context.Background(),
		layoutapi.DeleteInput{ID: fixtureRevisionID.String(), ExpectedRowRevision: 3})
	if err != nil {
		t.Fatal(err)
	}
}
