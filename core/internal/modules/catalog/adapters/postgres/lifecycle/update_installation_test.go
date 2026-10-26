package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

func TestUpdateRejectsEmptyInstallation(t *testing.T) {
	_, err := (&store{}).UpdateInstallation(context.Background(), catalogapp.InstallationMutation{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("update installation error = %v", err)
	}
}
