package folder

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestListBoundPreservesKeysetPage(t *testing.T) {
	store := fixtureStore{list: func(_ context.Context, _, _ domain.ID, limit uint16) (Page, error) {
		if limit != 25 {
			t.Fatal("unexpected limit")
		}
		return Page{Folders: []domain.Folder{fixtureFolderValue(1)}, NextID: fixtureFolder}, nil
	}}
	result, err := listBound(context.Background(), store, ListRequest{Limit: 25})
	if err != nil || len(result.Folders) != 1 || result.NextID != fixtureFolder {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
