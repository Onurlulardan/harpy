package adoption

import (
	"context"
	"errors"
	"testing"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

func TestFinalizeRunRejectsEmptyRun(t *testing.T) {
	_, err := (&store{}).FinalizeRun(context.Background(), adoptionapp.Run{})
	if !errors.Is(err, adoptionapp.ErrInvalid) {
		t.Fatalf("finalize adoption run error = %v", err)
	}
}
