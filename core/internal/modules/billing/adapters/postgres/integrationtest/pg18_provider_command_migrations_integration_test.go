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

type providerPGDocument struct {
	path    string
	content []byte
}

func applyProviderPGDocuments(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	asyncDirectory := filepath.Join(
		workingDirectory, "..", "..", "..", "..", "..", "platform", "async",
		"migrations", "fresh",
	)
	billingDirectory := filepath.Join(
		workingDirectory, "..", "..", "..", "migrations", "platform", "fresh",
	)
	applyProviderPGDirectory(t, ctx, database, asyncDirectory, true)
	applyProviderPGDirectory(t, ctx, database, billingDirectory, false)
}

func applyProviderPGDirectory(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	directory string,
	async bool,
) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	documents := make([]providerPGDocument, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sql") ||
			(async && !providerPGAsyncDocument(entry.Name())) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		documents = append(documents, providerPGDocument{path: path, content: content})
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].path < documents[right].path
	})
	if async && len(documents) != 8 {
		t.Fatalf("provider async prerequisite count=%d want=8", len(documents))
	}
	for _, document := range documents {
		if _, err := database.ExecContext(ctx, string(document.content)); err != nil {
			t.Fatalf("apply %s: %v", filepath.Base(document.path), err)
		}
	}
}

func providerPGAsyncDocument(name string) bool {
	for _, prefix := range []string{
		"20260715020000_", "20260715020100_", "20260715020200_",
		"20260715020300_", "20260715020400_", "20260715020500_",
		"20260715020600_", "20260715022010_",
	} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
