//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"strings"
	"testing"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

const (
	layoutFirstRevision   = "aaaaaaaa-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	layoutSecondRevision  = "bbbbbbbb-bbbb-7bbb-8bbb-bbbbbbbbbbbb"
	layoutDeletedRevision = "cccccccc-cccc-7ccc-8ccc-cccccccccccc"
	layoutCatalog         = `{"contract":"forms-presentation-field-catalog.v1","fields":[` +
		`{"id":"33333333-3333-4333-8333-333333333333","mode":"editable",` +
		`"type_id":"text","value_profile":"text_nfc_v1"},` +
		`{"id":"44444444-4444-4444-8444-444444444444","mode":"editable",` +
		`"type_id":"text","value_profile":"text_nfc_v1"}]}`
	layoutPresentation = `{"default_locale":"en","locales":["en"],` +
		`"registry_contract":"forms-presentation-registry.v1","surfaces":[{` +
		`"id":"018f0f4a-0000-7000-8000-000000000001","kind":"record_create",` +
		`"root":{"children":[{"children":[{"field_id":` +
		`"33333333-3333-4333-8333-333333333333","id":` +
		`"018f0f4a-0000-7000-8000-000000000004","interaction":"auto",` +
		`"kind":"field","span":1,"translations":[{"label":"Name",` +
		`"locale":"en"}],"widget":"text"}],"id":` +
		`"018f0f4a-0000-7000-8000-000000000003","kind":"container",` +
		`"span":1,"translations":[{"label":"General","locale":"en"}],` +
		`"variant":"section"}],"id":"018f0f4a-0000-7000-8000-000000000002",` +
		`"kind":"container","span":1,"translations":[],"variant":"stack"},` +
		`"translations":[{"label":"Create","locale":"en"}]}]}`
)

func assertLayoutCRUDAndPublication(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	auto := adapterDefinition
	text := strings.Replace(auto, `"widget":"auto"`, `"widget":"text"`, 1)
	created := createLayoutRevision(t, ctx, database, "layout_a",
		layoutFirstRevision, layoutapi.ResourceUser, auto)
	if created.Version != 1 || created.State != layoutapi.RevisionDraft ||
		created.RowRevision != 1 {
		t.Fatalf("created layout = %+v", created)
	}
	updated := updateLayoutRevision(t, ctx, database, "layout_a",
		layoutFirstRevision, 1, text)
	if updated.RowRevision != 2 ||
		updated.Definition.Surfaces[0].Sections[0].Fields[0].Widget !=
			layoutapi.WidgetText {
		t.Fatalf("updated layout = %+v", updated)
	}
	_, err := runLayoutRuntimeText(ctx, database, "layout_a", updateSQL,
		layoutFirstRevision, 1, text, layoutDigest(text), layoutNow(3))
	assertLayoutSQLState(t, err, "40001")
	invalidArguments := layoutPublishArguments(
		layoutFirstRevision, 2, text, layoutNow(3),
	)
	invalidCatalog := `{"contract":"forms-presentation-field-catalog.v1","fields":[]}`
	invalidArguments[5] = invalidCatalog
	invalidArguments[6] = layoutDigest(invalidCatalog)
	_, err = runLayoutRuntimeText(
		ctx, database, "layout_a", publishSQL, invalidArguments...,
	)
	assertLayoutSQLState(t, err, "23514")
	draft := readLayoutRevision(t, ctx, database, "layout_a", getSQL,
		layoutFirstRevision)
	if draft.State != layoutapi.RevisionDraft || draft.RowRevision != 2 {
		t.Fatalf("failed publish changed draft = %+v", draft)
	}
	published := publishLayoutRevision(t, ctx, database, "layout_a",
		layoutFirstRevision, 2, text, layoutNow(4))
	if published.Revision.State != layoutapi.RevisionPublished ||
		published.Revision.RowRevision != 3 ||
		published.Revision.PublishedAt == nil || !published.EvidenceComplete {
		t.Fatalf("published layout = %+v", published)
	}
	assertPublishedLayoutIsTerminal(t, ctx, database, text)
	second := createLayoutRevision(t, ctx, database, "layout_a",
		layoutSecondRevision, layoutapi.ResourceUser, auto)
	if second.Version != 2 {
		t.Fatalf("second version = %d", second.Version)
	}
	publishLayoutRevision(t, ctx, database, "layout_a", layoutSecondRevision, 1,
		auto, layoutNow(6))
	resolved := readPublishedLayout(t, ctx, database, "layout_a", resolveSQL, "user")
	if resolved.Revision.ID != layoutSecondRevision || !resolved.EvidenceComplete {
		t.Fatalf("resolved publication = %+v", resolved)
	}
	assertLayoutList(t, ctx, database)
	createLayoutRevision(t, ctx, database, "layout_a", layoutDeletedRevision,
		layoutapi.ResourceRole, auto)
	deleted, err := runLayoutRuntimeText(ctx, database, "layout_a", deleteSQL,
		layoutDeletedRevision, 1)
	if err != nil || deleted != "true" {
		t.Fatalf("delete result = %q: %v", deleted, err)
	}
	_, err = runLayoutRuntimeText(ctx, database, "layout_a", getSQL,
		layoutDeletedRevision)
	assertLayoutSQLState(t, err, "P0002")
}

func createLayoutRevision(t *testing.T, ctx context.Context, database *sql.DB,
	schema, id string, resource layoutapi.ResourceKind, definition string,
) layoutapi.Revision {
	t.Helper()
	return readLayoutRevision(t, ctx, database, schema, createSQL, id,
		string(resource), definition, layoutDigest(definition), layoutNow(1))
}

func updateLayoutRevision(t *testing.T, ctx context.Context, database *sql.DB,
	schema, id string, expected int64, definition string,
) layoutapi.Revision {
	t.Helper()
	return readLayoutRevision(t, ctx, database, schema, updateSQL, id, expected,
		definition, layoutDigest(definition), layoutNow(expected+1))
}

func publishLayoutRevision(t *testing.T, ctx context.Context, database *sql.DB,
	schema, id string, expected int64, definition string, now time.Time,
) layoutapp.PublishedPresentation {
	t.Helper()
	return readPublishedLayout(
		t, ctx, database, schema, publishSQL,
		layoutPublishArguments(id, expected, definition, now)...,
	)
}

func layoutPublishArguments(
	id string,
	expected int64,
	definition string,
	now time.Time,
) []any {
	return []any{
		id, expected, layoutDigest(definition),
		"system-resource-presentation-compiler.v2",
		"forms-presentation-field-catalog.v1",
		layoutCatalog, layoutDigest(layoutCatalog),
		"forms-presentation.v1",
		layoutPresentation, layoutDigest(layoutPresentation), now,
	}
}

func readPublishedLayout(t *testing.T, ctx context.Context, database *sql.DB,
	schema, statement string, arguments ...any,
) layoutapp.PublishedPresentation {
	t.Helper()
	document, err := runLayoutRuntimeText(ctx, database, schema, statement, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	published, err := decodePublishedPresentation([]byte(document))
	if err != nil {
		t.Fatalf("decode runtime publication: %v", err)
	}
	return published
}

func readLayoutRevision(t *testing.T, ctx context.Context, database *sql.DB,
	schema, statement string, arguments ...any,
) layoutapi.Revision {
	t.Helper()
	document, err := runLayoutRuntimeText(ctx, database, schema, statement, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := decodeRevision([]byte(document))
	if err != nil {
		t.Fatalf("decode runtime revision: %v", err)
	}
	return revision
}

func layoutDigest(definition string) []byte {
	digest := sha256.Sum256([]byte(definition))
	return digest[:]
}

func layoutNow(offset int64) time.Time {
	return adapterNow.Add(time.Duration(offset) * time.Minute)
}
