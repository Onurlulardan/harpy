package folder

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestRestoreBoundReturnsActiveRevision(t *testing.T) {
	store := fixtureStore{restore: func(context.Context, Mutation) (domain.Folder, bool, error) {
		return fixtureFolderValue(6), false, nil
	}}
	result, err := restoreBound(context.Background(), store, fixtureMutation(5))
	if err != nil || result.State != domain.FolderActive || result.RowVersion != 6 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
