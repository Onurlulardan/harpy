//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type recoveryAnchorDocument struct {
	path    string
	content []byte
}

func applyRecoveryAnchorMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	prepareRecoveryAnchorDatabase(t, ctx, database)
	root := recoveryAnchorRepositoryRoot(t)
	platformOps := loadRecoveryAnchorDocuments(t, root,
		"core/internal/modules/platformops/migrations/platform", origin, "")
	applyOriginDocuments(t, ctx, database, platformOps, origin, "")
	prepareRecoveryAnchorPrerequisites(t, ctx, database)
	iamDocuments := loadRecoveryAnchorDocuments(t, root,
		"core/internal/modules/iam/migrations/platform", origin, "2026072306000")
	applyOriginDocuments(t, ctx, database, iamDocuments, origin, "2026072306000")
}

func applyOriginDocuments(
	t *testing.T, ctx context.Context, database *sql.DB,
	documents []recoveryAnchorDocument, origin, versionPrefix string,
) {
	t.Helper()
	needle := "/migrations/platform/" + origin + "/"
	applied := 0
	for _, document := range documents {
		if !strings.Contains(document.path, needle) ||
			(versionPrefix != "" && !strings.Contains(document.path, versionPrefix)) {
			continue
		}
		if _, err := database.ExecContext(ctx, string(document.content)); err != nil {
			t.Fatalf("apply %s: %v", document.path, err)
		}
		applied++
	}
	if versionPrefix != "" && applied != 6 {
		t.Fatalf("recovery anchor migration count = %d, want 6", applied)
	}
}

func loadRecoveryAnchorDocuments(
	t *testing.T, root, relative, origin, prefix string,
) []recoveryAnchorDocument {
	t.Helper()
	directory := filepath.Join(root, filepath.FromSlash(relative), origin)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	documents := make([]recoveryAnchorDocument, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sql") ||
			(prefix != "" && !strings.HasPrefix(entry.Name(), prefix)) {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		documents = append(documents, recoveryAnchorDocument{
			path: filepath.ToSlash(path), content: content,
		})
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].path < documents[right].path
	})
	return documents
}

func recoveryAnchorRepositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for range 12 {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		directory = filepath.Dir(directory)
	}
	t.Fatal("repository root is missing")
	return ""
}
