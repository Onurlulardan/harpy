package v1adoption

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCutoverMigrationsArePairedBoundedAndSealed(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "tenant")
	fresh := workCutoverPaths(t, filepath.Join(root, "fresh"), ".fresh.sql")
	adopted := workCutoverPaths(t, filepath.Join(root, "v1_adopted"), ".v1_adopted.sql")
	if len(fresh) != 14 || len(adopted) != len(fresh) {
		t.Fatalf("Work cutover migration count fresh=%d adopted=%d", len(fresh), len(adopted))
	}
	for index := range fresh {
		left := readWorkCutoverMigration(t, fresh[index])
		right := readWorkCutoverMigration(t, adopted[index])
		if !bytes.Equal(left, right) {
			t.Fatalf("paired Work cutover migration differs: %s", filepath.Base(fresh[index]))
		}
		if lines := bytes.Count(left, []byte("\n")) + 1; lines > 200 {
			t.Fatalf("Work cutover migration has %d lines: %s", lines, filepath.Base(fresh[index]))
		}
	}
}

func TestCutoverCompositeUsesGenericReceiptsAndExactGrant(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "tenant", "fresh")
	composite := string(readWorkCutoverMigration(t, filepath.Join(root,
		"20260718230131_create_work_cutover_composite_v1.fresh.sql")))
	for _, token := range []string{
		"platform.cutover_prepare_batch_v1", "platform.cutover_record_batch_v1",
		"work_v1_adoption_apply_v1", "SESSION_USER", "SET search_path FROM CURRENT",
		"work.v1_adoption.v1",
	} {
		if !strings.Contains(composite, token) {
			t.Fatalf("Work cutover composite missing %q", token)
		}
	}
	grant := string(readWorkCutoverMigration(t, filepath.Join(root,
		"20260718230132_grant_work_cutover_v1.fresh.sql")))
	if strings.Count(grant, "GRANT EXECUTE ON FUNCTION") != 1 ||
		!strings.Contains(grant, "work_v1_adoption_apply_and_record_v1") ||
		strings.Contains(grant, "work_v1_adoption_apply_v1") {
		t.Fatalf("Work adopter grant is wider than the composite: %s", grant)
	}
}

func TestCutoverApplyEmitsExactChangeReceipts(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "tenant", "fresh")
	apply := string(readWorkCutoverMigration(t, filepath.Join(root,
		"20260718230125_create_work_cutover_apply_v1.fresh.sql")))
	for _, token := range []string{
		"platform.applied_change_receipts", "operation.lease_manifest_sha256",
		"(p_batch_ordinal - 1)::bigint * 1000", "p_batch_sha256 || ':'",
	} {
		if !strings.Contains(apply, token) {
			t.Fatalf("Work cutover receipt apply missing %q", token)
		}
	}
}

func TestCutoverEvidenceGuardsQuietBarrierAndStageDigest(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "tenant", "fresh")
	batch := string(readWorkCutoverMigration(t, filepath.Join(root,
		"20260718230124_create_work_cutover_batch_assert_v1.fresh.sql")))
	if !strings.Contains(batch, "p_final_lsn IS DISTINCT FROM p_commit_lsn") {
		t.Fatal("Work quiet completion does not bind final LSN to commit LSN")
	}
	stage := string(readWorkCutoverMigration(t, filepath.Join(root,
		"20260718230135_add_work_cutover_stage_digest_guard.fresh.sql")))
	if !strings.Contains(stage, "target_sha256 = pg_catalog.sha256") {
		t.Fatal("Work stage target digest constraint is missing")
	}
}

func TestCutoverRevokesLegacyRuntimeAdoptionWriters(t *testing.T) {
	path := filepath.Join("..", "..", "..", "migrations", "tenant", "fresh",
		"20260718230120_create_work_cutover_stage_v1.fresh.sql")
	content := string(readWorkCutoverMigration(t, path))
	for _, table := range []string{
		"work_v1_writer_fence", "work_adoption_runs", "work_adoption_objects",
		"work_adoption_issues", "work_reverse_reconciliations",
	} {
		if !strings.Contains(content, "ON "+table+" FROM harpy_runtime") {
			t.Fatalf("Work runtime revocation missing for %s", table)
		}
	}
}

func workCutoverPaths(t *testing.T, directory, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 0, 14)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "202607182301") &&
			strings.HasSuffix(entry.Name(), suffix) {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	return paths
}

func readWorkCutoverMigration(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Work cutover migration %s: %v", path, err)
	}
	return content
}
