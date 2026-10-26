package adoption

import (
	"context"
	"errors"
	"testing"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

func TestRetireRejectsEmptyRoute(t *testing.T) {
	err := (&store{}).Retire(context.Background(), "", time.Time{})
	if !errors.Is(err, adoptionapp.ErrInvalid) {
		t.Fatalf("retire compatibility edge error = %v", err)
	}
}
