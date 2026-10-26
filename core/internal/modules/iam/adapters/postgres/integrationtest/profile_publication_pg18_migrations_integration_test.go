//go:build integration

package integrationtest

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func applyPublicationRuntimeAuthorityDocuments(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	origin := db.tenant.origin
	cutoverName := "20260718050601_create_cutover_runtime_locks_v1." + origin + ".sql"
	documents := make([]migration.Document, 0, 5)
	for _, base := range []string{
		"20260718050600_create_tenancy_runtime_resolver_v1",
		"20260725125320_create_tenancy_runtime_identity_resolver_v1",
		"20260725125340_create_tenancy_runtime_identity_resolver_v2",
		"20260725125350_create_tenancy_runtime_identity_resolver_v3",
	} {
		name := base + "." + origin + ".sql"
		documents = append(documents, loadPublicationAuthorityDocument(t, "tenancy",
			"core/internal/modules/tenancy/migrations/platform/"+origin+"/"+name,
			filepath.Join("..", "..", "..", "..", "tenancy", "migrations",
				"platform", origin, name)))
	}
	documents = append(documents,
		loadPublicationAuthorityDocument(t, "platform_cutover",
			"core/internal/platform/cutover/migrations/"+origin+"/"+cutoverName,
			filepath.Join("..", "..", "..", "..", "..", "platform", "cutover",
				"migrations", origin, cutoverName)))
	applyPublicationDocuments(t, ctx, db, documents, false)
}

func loadPublicationAuthorityDocument(
	t *testing.T, owner, logicalPath, diskPath string,
) migration.Document {
	t.Helper()
	content, err := os.ReadFile(diskPath)
	if err != nil {
		t.Fatalf("read publication authority migration %s: %v", logicalPath, err)
	}
	document, err := migration.NewDocument(owner, "platform", logicalPath, content)
	if err != nil {
		t.Fatalf("bind publication authority migration %s: %v", logicalPath, err)
	}
	return document
}

func publicationIAMDocuments(t *testing.T, origin, scope, phase string) []migration.Document {
	t.Helper()
	documents, err := iammodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	return selectPublicationDocuments(t, documents, origin, scope, func(version int64) bool {
		switch phase {
		case "foundation":
			return scope == "tenant" && version >= 20260716011000 &&
				version <= 20260716012900
		case "platform":
			return scope == "platform" &&
				(version == 20260716010000 || version == 20260716012020 ||
					version >= 20260716013705 && version <= 20260716013707 ||
					version == 20260716014600 || version == 20260716014601)
		case "runtime":
			return scope == "tenant" && publicationRuntimeVersion(version)
		default:
			t.Fatalf("unknown publication migration phase %q", phase)
			return false
		}
	})
}

func publicationRuntimeVersion(version int64) bool {
	return version >= 20260716012920 && version <= 20260716012926 ||
		version >= 20260716013020 && version <= 20260716013220 ||
		version >= 20260716013424 && version <= 20260716013425 ||
		version == 20260716013433 ||
		version >= 20260716013700 && version <= 20260716013806 ||
		version >= 20260716014300 && version <= 20260716014359 ||
		version == 20260716014612 || version == 20260716014658 ||
		version == 20260716015209 || version == 20260716015212 ||
		version >= 20260716015910 && version <= 20260716015954 ||
		version >= 20260716020002 && version <= 20260716020101
}

func publicationAsyncDocuments(t *testing.T, origin string) []migration.Document {
	t.Helper()
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	return selectPublicationDocuments(t, documents, origin, "platform", func(version int64) bool {
		return version <= 20260715022010
	})
}

func selectPublicationDocuments(
	t *testing.T,
	documents []migration.Document,
	origin, scope string,
	selected func(int64) bool,
) []migration.Document {
	t.Helper()
	result := make([]migration.Document, 0)
	needle := "/" + origin + "/"
	for _, document := range documents {
		path := filepath.ToSlash(document.Path())
		if document.Scope() != scope || !strings.Contains(path, needle) {
			continue
		}
		base := filepath.Base(path)
		version, err := strconv.ParseInt(base[:14], 10, 64)
		if err != nil {
			t.Fatalf("parse migration %s: %v", path, err)
		}
		if selected(version) {
			result = append(result, document)
		}
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Path() < result[right].Path()
	})
	if len(result) == 0 {
		t.Fatalf("no %s/%s publication migrations selected", scope, origin)
	}
	return result
}

func applyPublicationDocuments(
	t *testing.T,
	ctx context.Context,
	db publicationDatabase,
	documents []migration.Document,
	tenant bool,
) {
	t.Helper()
	for _, document := range documents {
		tx, err := db.admin.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_owner")
		if err == nil {
			path := "platform,pg_catalog"
			if tenant {
				path = db.tenant.schema + ",platform,pg_catalog"
			}
			_, err = tx.Exec(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", path)
		}
		if err == nil {
			_, err = tx.Exec(ctx, string(document.Content()))
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatalf("commit %s: %v", document.Path(), err)
		}
	}
}
