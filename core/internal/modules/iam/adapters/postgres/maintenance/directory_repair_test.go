package maintenance

import (
	"context"
	"errors"
	"testing"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func TestExecuteDirectoryRepairRejectsNilStore(t *testing.T) {
	var store *store
	_, err := store.executeDirectoryRepair(context.Background(), maintenanceapp.Claim{}, uuid.Nil)
	if !errors.Is(err, ErrRuntimeInvalid) {
		t.Fatalf("nil store error = %v", err)
	}
}
