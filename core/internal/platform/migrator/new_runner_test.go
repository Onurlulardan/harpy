package migrator

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func TestNewRunnerRejectsUnverifiedAuthority(t *testing.T) {
	_, err := NewRunner(RunnerDependencies{
		Database:       nil,
		CleanupTimeout: requiredCleanupTimeout,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err == nil {
		t.Fatal("unverified runner dependencies were accepted")
	}
}

func TestNewRunnerRequiresEmbeddedReleaseBindingBeforeArtifacts(t *testing.T) {
	_, err := NewRunner(RunnerDependencies{
		Database:       &sql.DB{},
		Registry:       &migration.ReleaseRegistry{},
		CleanupTimeout: requiredCleanupTimeout,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err == nil || !strings.Contains(err.Error(), "embedded schema contract") {
		t.Fatalf("missing embedded release binding error = %v", err)
	}
}
