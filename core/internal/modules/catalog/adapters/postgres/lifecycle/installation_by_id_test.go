package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

func TestInstallationByIDRejectsNonV7Identity(t *testing.T) {
	_, err := (&store{}).InstallationByID(context.Background(), uuid.New())
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("get installation by id error = %v", err)
	}
}
