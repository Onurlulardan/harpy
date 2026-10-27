package menu

import (
	"context"
	"errors"
	"testing"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func TestUpdateRejectsEmptyMutation(t *testing.T) {
	_, err := (&store{}).Update(context.Background(), menuapp.OverrideMutation{})
	if !errors.Is(err, menuapp.ErrInvalid) {
		t.Fatalf("update menu override error = %v", err)
	}
}
