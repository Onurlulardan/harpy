package lifecycle

import (
	"context"
	"errors"
	"strings"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

func TestListRejectsUnboundedInstallations(t *testing.T) {
	_, err := (&store{}).ListInstallations(context.Background(), catalogapp.PageRequest{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("list installations error = %v", err)
	}
}

func TestListInstallationsSQLProjectsLastOperation(t *testing.T) {
	if !strings.Contains(listInstallationsSQL, "catalog_operation_runs") ||
		!strings.Contains(listInstallationsSQL, "last_operation.operation_id") {
		t.Fatal("installation list omitted last operation identity")
	}
}
