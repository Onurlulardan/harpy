//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCatalogPG18ProtectedCutoverFreshAndAdoptedEvidence(t *testing.T) {
	t.Run("fresh-denied", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		database := startCatalogPG18(t, ctx, "fresh-cutover")
		prepareCatalogPG18(t, ctx, database, "fresh")
		assertCatalogMenuBaselineState(t, ctx, database.owner, "catalog_a", 1)
		authority := prepareCatalogCutoverAuthority(t, ctx, database.owner, "fresh")
		evidence := newCatalogCutoverEvidence(nil, nil)
		batch := newCatalogCutoverBatch(authority, "modules", 1, evidence, nil)
		if _, _, _, err := executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, batch,
		); err == nil {
			t.Fatal("fresh-origin Catalog V1 adoption was accepted")
		}
		assertCatalogCutoverCounts(t, ctx, database.owner, authority, 0, 0, 0, 0)
	})

	t.Run("v1-adopted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		database := startCatalogPG18(t, ctx, "v1-adopted-cutover")
		prepareCatalogPG18InactiveAdoption(t, ctx, database)
		assertCatalogMenuBaselineState(t, ctx, database.owner, "catalog_a", 0)
		assertCatalogMenuRepairRejectsTamper(t, ctx, database.owner)
		assertCatalogMenuRepairRejectsAuthorityMismatch(t, ctx, database.owner)
		authority := prepareCatalogCutoverAuthority(
			t, ctx, database.owner, "v1_adopted",
		)
		mappings := loadCatalogCutoverMappings(t, ctx, database.owner, authority)
		consumerRows := seedCatalogCutoverConsumer(t, ctx, database.owner)
		evidence := newCatalogCutoverEvidence(mappings, consumerRows)
		batches := buildCatalogCutoverBatches(authority, evidence, mappings)

		count, keyDigest, stateDigest, err := executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, batches[0],
		)
		if err != nil || count != 1 || len(keyDigest) != 64 || len(stateDigest) != 64 {
			t.Fatalf("Catalog protected first batch=%d %q %q %v",
				count, keyDigest, stateDigest, err)
		}
		replayCount, replayKey, replayState, err := executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, batches[0],
		)
		if err != nil || replayCount != count || replayKey != keyDigest ||
			replayState != stateDigest {
			t.Fatalf("Catalog exact replay=%d %q %q %v",
				replayCount, replayKey, replayState, err)
		}
		assertCatalogCutoverTamperAndIsolation(
			t, ctx, database.owner, authority, evidence, mappings[0],
		)
		for index := 1; index < len(batches)-1; index++ {
			if _, _, _, err = executeCatalogCutoverBatch(
				ctx, database.owner, authority, authority.schema, batches[index],
			); err != nil {
				t.Fatalf("Catalog protected batch %d: %v", index+1, err)
			}
		}
		final := batches[len(batches)-1]
		blockCatalogCutoverInventory(t, ctx, database.owner)
		if _, _, _, err = executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, final,
		); err == nil || !strings.Contains(err.Error(), "inventory evidence") {
			t.Fatalf("Catalog blocker denial=%v", err)
		}
		unblockCatalogCutoverInventory(t, ctx, database.owner)
		unretireCatalogCutoverConsumer(t, ctx, database.owner)
		if _, _, _, err = executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, final,
		); err == nil || !strings.Contains(err.Error(), "consumers are not retired") {
			t.Fatalf("Catalog consumer retirement denial=%v", err)
		}
		retireCatalogCutoverConsumer(t, ctx, database.owner)
		if count, _, stateDigest, err = executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, final,
		); err != nil || count != 3 || stateDigest != evidence.TargetRootSHA256 {
			t.Fatalf("Catalog protected final batch=%d %q %v", count, stateDigest, err)
		}
		if replayCount, replayKey, replayState, err = executeCatalogCutoverBatch(
			ctx, database.owner, authority, authority.schema, final,
		); err != nil || replayCount != 3 || replayState != evidence.TargetRootSHA256 ||
			len(replayKey) != 64 {
			t.Fatalf("Catalog final exact replay=%d %q %q %v",
				replayCount, replayKey, replayState, err)
		}
		assertCatalogCutoverCounts(t, ctx, database.owner, authority, 5, 6, 1, 1)
		assertCatalogCutoverEvidence(t, ctx, database.owner, authority, evidence)
		assertCatalogCutoverMenuPreserved(t, ctx, database.owner)
		assertCatalogCutoverDirectWritesDenied(t, ctx, database.owner, authority)
	})
}

func buildCatalogCutoverBatches(authority catalogCutoverAuthority,
	evidence catalogCutoverEvidence, mappings []catalogCutoverMapping,
) []catalogCutoverBatch {
	byCheckpoint := map[string][]catalogCutoverMapping{}
	for index := range mappings {
		key := mappings[index].checkpoint
		byCheckpoint[key] = append(byCheckpoint[key], mappings[index])
	}
	keys := []string{"modules", "module_dependencies", "module_forms",
		"module_schema_versions", "module_migration_runs", "menu_configs"}
	result := make([]catalogCutoverBatch, 0, len(keys))
	for index, key := range keys {
		result = append(result, newCatalogCutoverBatch(
			authority, key, index+1, evidence, byCheckpoint[key],
		))
	}
	return result
}

func assertCatalogCutoverTamperAndIsolation(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority,
	evidence catalogCutoverEvidence, mapping catalogCutoverMapping,
) {
	t.Helper()
	var target map[string]any
	if err := json.Unmarshal([]byte(mapping.targetRecord), &target); err != nil {
		t.Fatal(err)
	}
	target["active_version"] = "1.0.1"
	tampered, err := json.Marshal(target)
	if err != nil {
		t.Fatal(err)
	}
	mapping.targetRecord = string(tampered)
	batch := newCatalogCutoverBatch(
		authority, "modules", 1, evidence, []catalogCutoverMapping{mapping},
	)
	if _, _, _, err = executeCatalogCutoverBatch(
		ctx, database, authority, authority.schema, batch,
	); err == nil {
		t.Fatal("Catalog different-hash replay was accepted")
	}
	if _, _, _, err = executeCatalogCutoverBatch(
		ctx, database, authority, "catalog_b", batch,
	); err == nil {
		t.Fatal("Catalog cross-tenant apply was accepted")
	}
	var stage int
	if err = database.QueryRowContext(ctx,
		"SELECT count(*) FROM catalog_b.catalog_v1_adoption_stage").Scan(&stage); err != nil || stage != 0 {
		t.Fatalf("Catalog cross-tenant stage=%d err=%v", stage, err)
	}
}
