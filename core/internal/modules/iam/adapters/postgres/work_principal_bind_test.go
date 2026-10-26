package postgres

import (
	"context"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestWorkPrincipalReaderRejectsUnownedTenantScope(t *testing.T) {
	reader, err := NewWorkPrincipalReader()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.Evidence(
		context.Background(), platformtx.TenantScope{}, "USER", uuid.New(),
	); err == nil {
		t.Fatal("principal evidence accepted unowned tenant scope")
	}
	if _, err := reader.AvailableUsers(
		context.Background(), platformtx.TenantScope{}, "", 25,
	); err == nil {
		t.Fatal("principal directory accepted unowned tenant scope")
	}
}

func TestWorkPrincipalReaderRejectsUnknownTargetKind(t *testing.T) {
	reader, err := NewWorkPrincipalReader()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.Evidence(
		context.Background(), platformtx.TenantScope{}, "TASK", uuid.New(),
	); err == nil {
		t.Fatal("principal evidence accepted unknown target kind")
	}
}
