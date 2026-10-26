package menu

import (
	"errors"
	"testing"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func TestReplaceRejectsEmptyMutation(t *testing.T) {
	_, err := (&store{}).Replace(nil, menuapp.ReplaceMutation{})
	if !errors.Is(err, menuapp.ErrInvalid) {
		t.Fatalf("replace menu error = %v", err)
	}
}
