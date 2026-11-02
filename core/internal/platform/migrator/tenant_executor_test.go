package migrator

import (
	"database/sql"
	"io"
	"log/slog"
	"testing"
	"testing/fstest"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

func TestNewTenantExecutorRejectsUnverifiedRelease(t *testing.T) {
	_, err := NewTenantExecutor(TenantExecutorDependencies{
		Database:   &sql.DB{},
		OpenTenant: func() (*sql.DB, error) { return &sql.DB{}, nil },
		Release:    attestation.VerifiedRelease{},
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		NewID:      func() (uuid.UUID, error) { return uuid.New(), nil },
	})
	if err == nil {
		t.Fatal("unverified release built a tenant executor")
	}
}

func TestTenantBaselineSelectionBuildsTarget(t *testing.T) {
	selected := tenantBaselineSelection()
	target, err := NewTenantTargetExecutor(TenantTargetDependencies{
		Database: func() (*sql.DB, error) { return &sql.DB{}, nil },
		Files:    fstest.MapFS{},
		Selected: []postgresv1.SelectedMigration{selected},
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil || target == nil || target.runner == nil {
		t.Fatalf("baseline target: %v", err)
	}
	if selected.Version != 20261028000001 || selected.Scope != tenantMigrationScope {
		t.Fatalf("baseline selection = %+v", selected)
	}
}
