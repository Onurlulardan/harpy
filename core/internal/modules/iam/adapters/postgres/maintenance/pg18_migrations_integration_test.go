//go:build integration

package maintenance

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

type maintenancePG18Document struct {
	path    string
	content []byte
}

func loadMaintenancePG18Migrations(
	t *testing.T, origin, scope string, selected func(int64) bool, want int,
) []maintenancePG18Document {
	t.Helper()
	if origin != "fresh" {
		t.Fatalf("unsupported maintenance origin: %s", origin)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(
		workingDirectory, "..", "..", "..", "migrations", scope, origin,
	)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read maintenance migrations: %v", err)
	}
	documents := make([]maintenancePG18Document, 0, want)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, parseErr := strconv.ParseInt(entry.Name()[:14], 10, 64)
		if parseErr != nil || !selected(version) {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		documents = append(documents, maintenancePG18Document{path, content})
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].path < documents[right].path
	})
	if len(documents) != want {
		t.Fatalf("%s/%s migration count=%d, want %d", origin, scope, len(documents), want)
	}
	return documents
}

func applyMaintenancePG18Platform(
	t *testing.T, ctx context.Context, database *sql.DB,
	documents []maintenancePG18Document,
) {
	t.Helper()
	for _, document := range documents {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = transaction.ExecContext(ctx, string(document.content)); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s: %v", document.path, err)
		}
		if err = transaction.Commit(); err != nil {
			t.Fatalf("commit %s: %v", document.path, err)
		}
	}
}

func applyMaintenancePG18Tenant(
	t *testing.T, ctx context.Context, database *sql.DB,
	documents []maintenancePG18Document,
) {
	t.Helper()
	for _, tenant := range maintenancePG18Tenants {
		for _, document := range documents {
			transaction, err := database.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = transaction.ExecContext(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)",
				tenant.schema+", pg_catalog")
			if err == nil {
				_, err = transaction.ExecContext(ctx, string(document.content))
			}
			if err != nil {
				_ = transaction.Rollback()
				t.Fatalf("apply %s to %s: %v", document.path, tenant.schema, err)
			}
			if err = transaction.Commit(); err != nil {
				t.Fatalf("commit %s for %s: %v", document.path, tenant.schema, err)
			}
		}
	}
}

func maintenanceFoundationVersion(scope string) func(int64) bool {
	return func(version int64) bool {
		if scope == "platform" {
			return version >= 20260716010000 && version <= 20260716012010
		}
		return version >= 20260716011000 && version <= 20260716012900
	}
}

func maintenancePrerequisiteVersion(version int64) bool {
	return version >= 20260716013020 && version <= 20260716013220 ||
		version == 20260716013433 ||
		version >= 20260716013700 && version <= 20260716013702 ||
		version >= 20260716013710 && version <= 20260716013713 ||
		version == 20260716013720 || version == 20260716013723 ||
		version == 20260716013730 || version == 20260716013733 ||
		version == 20260716013800
}

func maintenanceGuardVersion(version int64) bool {
	return version >= 20260716012920 && version <= 20260716012926
}

func maintenanceTenantVersion(version int64) bool {
	return version >= 20260716014300 && version <= 20260716014359 ||
		version == 20260716014612 || version == 20260716014658 ||
		version >= 20260716015920 && version <= 20260716015946 ||
		version >= 20260716015950 && version <= 20260716015954 ||
		version >= 20260716020002 && version <= 20260716020101 ||
		version == 20260716015209 || version == 20260716015212 ||
		version == 20260716015911
}

func maintenancePlatformVersion(version int64) bool {
	return version == 20260716014602 ||
		version >= 20260716014607 && version <= 20260716014609
}
