package menu

import (
	"context"
	"errors"
	"testing"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func TestGetActiveConfigRejectsUnboundStore(t *testing.T) {
	_, err := (&store{}).GetActiveConfig(context.Background())
	if !errors.Is(err, menuapp.ErrUnavailable) {
		t.Fatalf("get active menu config error = %v", err)
	}
}
