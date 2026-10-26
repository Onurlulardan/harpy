package folder

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestGetBoundReturnsExactStableFolder(t *testing.T) {
	store := fixtureStore{get: func(_ context.Context, id domain.ID) (domain.Folder, error) {
		if id != fixtureFolder {
			t.Fatal("unexpected folder id")
		}
		return fixtureFolderValue(4), nil
	}}
	result, err := getBound(context.Background(), store, fixtureFolder)
	if err != nil || result.ID != fixtureFolder || result.RowVersion != 4 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
