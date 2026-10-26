package folder

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestMoveBoundRequiresAdvancedRevision(t *testing.T) {
	store := fixtureStore{move: func(context.Context, Mutation) (domain.Folder, bool, error) {
		return fixtureFolderValue(4), false, nil
	}}
	_, err := moveBound(context.Background(), store, fixtureMutation(4))
	if err != ErrUnavailable {
		t.Fatalf("error=%v", err)
	}
}
