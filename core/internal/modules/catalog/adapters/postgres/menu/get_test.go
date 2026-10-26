package menu

import (
	"context"
	"errors"
	"testing"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func TestGetRejectsUnboundStore(t *testing.T) {
	_, err := (&store{}).Get(context.Background())
	if !errors.Is(err, menuapp.ErrUnavailable) {
		t.Fatalf("get menu error = %v", err)
	}
}
