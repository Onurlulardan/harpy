package application

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func TestGetReturnsOwnedRevisionCopy(t *testing.T) {
	stored := fixtureRevision(t, layoutapi.RevisionDraft, 1)
	store := &fixtureStore{}
	store.get = func(_ context.Context, id uuid.UUID) (layoutapi.Revision, error) {
		if id != fixtureRevisionID {
			t.Fatal("unexpected revision id")
		}
		return stored, nil
	}
	result, err := fixtureApplication(t, store).Get(
		context.Background(), fixtureRevisionID.String(),
	)
	if err != nil {
		t.Fatal(err)
	}
	result.Definition.Surfaces[0].Sections[0].Fields[0].Widget = layoutapi.WidgetJSON
	if stored.Definition.Surfaces[0].Sections[0].Fields[0].Widget == layoutapi.WidgetJSON {
		t.Fatal("stored revision escaped application ownership")
	}
}

func TestGetRejectsStoreIdentityDrift(t *testing.T) {
	store := &fixtureStore{}
	store.get = func(
		_ context.Context, _ uuid.UUID,
	) (layoutapi.Revision, error) {
		revision := fixtureRevision(t, layoutapi.RevisionDraft, 1)
		revision.ID = fixtureFieldOne.String()
		return revision, nil
	}
	_, err := fixtureApplication(t, store).Get(
		context.Background(), fixtureRevisionID.String(),
	)
	if err != layoutapi.ErrUnavailable {
		t.Fatalf("get identity drift error = %v", err)
	}
}
