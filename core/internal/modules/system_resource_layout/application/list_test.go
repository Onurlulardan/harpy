package application

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestListOrdersRevisionsByVersion(t *testing.T) {
	store := &fixtureStore{}
	store.list = func(
		_ context.Context, _ layoutapi.ResourceKind,
	) ([]layoutapi.Revision, error) {
		draft := fixtureRevision(t, layoutapi.RevisionDraft, 2)
		draft.ID = fixtureFieldOne.String()
		return []layoutapi.Revision{
			draft,
			fixtureRevision(t, layoutapi.RevisionPublished, 1),
		}, nil
	}
	result, err := fixtureApplication(t, store).List(
		context.Background(), layoutapi.ResourceUser)
	if err != nil || len(result) != 2 ||
		result[0].Version != 1 || result[1].Version != 2 {
		t.Fatalf("list result = %+v: %v", result, err)
	}
}

func TestListRejectsDuplicateVersion(t *testing.T) {
	store := &fixtureStore{}
	store.list = func(
		_ context.Context, _ layoutapi.ResourceKind,
	) ([]layoutapi.Revision, error) {
		first := fixtureRevision(t, layoutapi.RevisionDraft, 1)
		second := fixtureRevision(t, layoutapi.RevisionDraft, 1)
		second.ID = fixtureFieldOne.String()
		return []layoutapi.Revision{first, second}, nil
	}
	_, err := fixtureApplication(t, store).List(
		context.Background(), layoutapi.ResourceUser)
	if err != layoutapi.ErrUnavailable {
		t.Fatalf("duplicate layout version error = %v", err)
	}
}
