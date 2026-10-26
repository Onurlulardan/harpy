package menu

import (
	"context"
	"errors"
	"testing"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func TestCreateConfigRejectsUnboundStore(t *testing.T) {
	_, err := (&store{}).CreateConfig(context.Background(), menuapp.ConfigMutation{})
	if !errors.Is(err, menuapp.ErrInvalid) {
		t.Fatalf("create menu config error = %v", err)
	}
}

func TestUpdateConfigRejectsUnboundStore(t *testing.T) {
	_, err := (&store{}).UpdateConfig(context.Background(), menuapp.ConfigMutation{})
	if !errors.Is(err, menuapp.ErrInvalid) {
		t.Fatalf("update menu config error = %v", err)
	}
}
