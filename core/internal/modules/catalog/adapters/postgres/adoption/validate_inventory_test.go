package adoption

import (
	"context"
	"errors"
	"testing"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

func TestValidateInventoryRejectsUnboundStore(t *testing.T) {
	_, err := (&store{}).ValidateInventory(context.Background())
	if !errors.Is(err, adoptionapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
