package folder

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestDeleteBoundReturnsTrashedRevision(t *testing.T) {
	store := fixtureStore{delete: func(context.Context, Mutation) (domain.Folder, bool, error) {
		value := fixtureFolderValue(5)
		value.State = domain.FolderTrashed
		return value, false, nil
	}}
	result, err := deleteBound(context.Background(), store, fixtureMutation(4))
	if err != nil || result.State != domain.FolderTrashed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
