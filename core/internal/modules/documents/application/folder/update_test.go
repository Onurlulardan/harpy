package folder

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestUpdateBoundPropagatesRevisionConflict(t *testing.T) {
	store := fixtureStore{update: func(context.Context, Mutation) (domain.Folder, bool, error) {
		return domain.Folder{}, false, ErrConflict
	}}
	_, err := updateBound(context.Background(), store, fixtureMutation(4))
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}
