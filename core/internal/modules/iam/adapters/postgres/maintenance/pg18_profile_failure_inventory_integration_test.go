//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

type maintenanceInvalidPlanWire struct {
	SourceSchemaID uuid.UUID `json:"source_schema_id"`
	Pages          []struct {
		After   *uuid.UUID `json:"after_resource_id"`
		Through uuid.UUID  `json:"through_resource_id"`
		Count   int        `json:"expected_count"`
		Source  string     `json:"source_sha256"`
	} `json:"pages"`
}

func assertMaintenancePG18InvalidSourceInventory(
	t *testing.T, ctx context.Context, database *sql.DB,
	source maintenancePlanFixture,
) {
	t.Helper()
	rows, err := database.QueryContext(ctx, `SELECT id,
profile_schema_version_id,revision,profile_sha256,profile_values::text
FROM `+source.tenant.schema+`.iam_users ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	ids := append(append([]string(nil), maintenanceProfileUsers...), maintenanceLiveUser)
	revisions := []int64{4, 5, 4, 1}
	index := 0
	for rows.Next() {
		var id, schemaID uuid.UUID
		var revision int64
		var digest []byte
		var document string
		if err = rows.Scan(&id, &schemaID, &revision, &digest, &document); err != nil {
			t.Fatal(err)
		}
		if index >= len(ids) {
			t.Fatalf("invalid-plan source has an extra row %s", id)
		}
		wantDocument, wantSHA := maintenanceInvalidSourceDocument(index)
		if id != uuid.MustParse(ids[index]) ||
			schemaID != source.targetSchema.ID() || revision != revisions[index] ||
			!bytes.Equal(digest, wantSHA[:]) {
			t.Fatalf("invalid-plan source[%d]=%s/%s/%d/%x/%s, want %s/%s/%d/%x/%s",
				index, id, schemaID, revision, digest, document,
				ids[index], source.targetSchema.ID(),
				revisions[index], wantSHA, wantDocument)
		}
		index++
	}
	if err = rows.Err(); err != nil || index != len(ids) {
		t.Fatalf("invalid-plan source inventory count=%d: %v", index, err)
	}
	assertMaintenancePG18InvalidSourcePage(t, ctx, database, source,
		nil, maintenanceUserOne, []int{0})
	after := maintenanceProfileUsers[0]
	assertMaintenancePG18InvalidSourcePage(t, ctx, database, source,
		&after, maintenanceLiveUser, []int{1, 2, 3})
}

func assertMaintenancePG18InvalidSourcePage(
	t *testing.T, ctx context.Context, database *sql.DB,
	source maintenancePlanFixture, after *string, through string, indexes []int,
) {
	t.Helper()
	parts := make([][]byte, 0, len(indexes)*4)
	ids := append(append([]string(nil), maintenanceProfileUsers...), maintenanceLiveUser)
	revisions := []string{"4", "5", "4", "1"}
	for _, index := range indexes {
		_, digest := maintenanceInvalidSourceDocument(index)
		parts = append(parts, []byte(ids[index]),
			[]byte(source.targetSchema.ID().String()), []byte(revisions[index]), digest[:])
	}
	want := maintenancePageDigest(parts)
	var count int
	var actual []byte
	err := database.QueryRowContext(ctx, `SELECT * FROM `+source.tenant.schema+
		`.iam_profile_backfill_source_page_v2('user',$1,$2,$3)`,
		source.targetSchema.ID(), after, through).Scan(&count, &actual)
	if err != nil || count != len(indexes) || !bytes.Equal(actual, want[:]) {
		t.Fatalf("invalid-plan source page %v-%s=%d/%x, want %d/%x: %v",
			after, through, count, actual, len(indexes), want, err)
	}
}

func maintenanceInvalidSourceDocument(index int) (string, [32]byte) {
	if index == len(maintenanceProfileUsers) {
		return maintenanceLiveCurrentDocument(0)
	}
	return maintenanceMixedTargetDocument(0, index)
}

func assertMaintenancePG18InvalidPlanWire(
	t *testing.T, ctx context.Context, database *sql.DB,
	source maintenancePlanFixture, document []byte,
) {
	t.Helper()
	var wire maintenanceInvalidPlanWire
	if err := json.Unmarshal(document, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.SourceSchemaID != source.targetSchema.ID() || len(wire.Pages) != 2 {
		t.Fatalf("invalid-plan wire source=%s pages=%d",
			wire.SourceSchemaID, len(wire.Pages))
	}
	for index, page := range wire.Pages {
		var count int
		var actual []byte
		err := database.QueryRowContext(ctx, `SELECT * FROM `+source.tenant.schema+
			`.iam_profile_backfill_source_page_v2('user',$1,$2,$3)`,
			wire.SourceSchemaID, page.After, page.Through).Scan(&count, &actual)
		want, decodeErr := hex.DecodeString(page.Source)
		if err != nil || decodeErr != nil || count != page.Count ||
			!bytes.Equal(actual, want) {
			t.Fatalf("invalid-plan wire page %d=%d/%x, want %d/%x: %v/%v",
				index+1, count, actual, page.Count, want, err, decodeErr)
		}
	}
}
