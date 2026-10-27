//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type billingDocument struct {
	path    string
	content []byte
}

func loadBillingDocuments(t *testing.T, origin string) []billingDocument {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("Billing migration origin %q is unsupported", origin)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(
		workingDirectory, "..", "..", "..", "migrations", "platform", "fresh",
	)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read Billing migrations: %v", err)
	}
	documents := make([]billingDocument, 0, 200)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		documents = append(documents, billingDocument{path: path, content: content})
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].path < documents[right].path
	})
	if len(documents) != 208 {
		t.Fatalf("Billing fresh migration count = %d, want 208", len(documents))
	}
	return documents
}

func applyBillingDocuments(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	documents []billingDocument,
) {
	t.Helper()
	for _, document := range documents {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = transaction.ExecContext(ctx, string(document.content)); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s: %v", filepath.Base(document.path), err)
		}
		if err = transaction.Commit(); err != nil {
			t.Fatalf("commit %s: %v", filepath.Base(document.path), err)
		}
	}
}

func billingDocumentsThrough(
	t *testing.T,
	documents []billingDocument,
	maxVersion int64,
) []billingDocument {
	t.Helper()
	limit := len(documents)
	for index, document := range documents {
		name := filepath.Base(document.path)
		version, err := strconv.ParseInt(name[:14], 10, 64)
		if err != nil {
			t.Fatalf("parse Billing migration version %s: %v", name, err)
		}
		if version > maxVersion {
			limit = index
			break
		}
	}
	return documents[:limit]
}
