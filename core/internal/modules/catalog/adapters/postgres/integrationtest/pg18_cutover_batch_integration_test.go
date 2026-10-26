//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"sort"

	catalogadoption "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/adoption"
)

type catalogRootEntry struct{ kind, identity, document string }

type catalogCutoverEvidence struct {
	BlockerSHA256    string `json:"blocker_sha256"`
	ConsumerSHA256   string `json:"consumer_sha256"`
	EvidenceSHA256   string `json:"evidence_sha256"`
	SourceCount      int64  `json:"source_count"`
	SourceRootSHA256 string `json:"source_root_sha256"`
	TargetCount      int64  `json:"target_count"`
	TargetRootSHA256 string `json:"target_root_sha256"`
}

type catalogCutoverBatch struct {
	checkpoint, sourceXID, digest, lastKey string
	ordinal                                int
	bytes                                  []byte
}

func newCatalogCutoverEvidence(mappings []catalogCutoverMapping,
	consumers []catalogRootEntry,
) catalogCutoverEvidence {
	sourceByKey := make(map[string]catalogRootEntry, len(mappings))
	target := make([]catalogRootEntry, 0, len(mappings))
	for _, mapping := range mappings {
		kind := map[string]string{"modules": "module",
			"module_schema_versions": "schema_version",
			"menu_configs":           "menu"}[mapping.checkpoint]
		sourceByKey[kind+"\x00"+mapping.sourceIdentity] = catalogRootEntry{
			kind, mapping.sourceIdentity, mapping.sourceDocument,
		}
		target = append(target, catalogRootEntry{mapping.targetKind,
			mapping.targetIdentity, mapping.targetRecord})
	}
	source := make([]catalogRootEntry, 0, len(sourceByKey))
	for _, entry := range sourceByKey {
		source = append(source, entry)
	}
	evidence := catalogCutoverEvidence{BlockerSHA256: catalogFramedRoot(nil),
		ConsumerSHA256: catalogFramedRoot(consumers), SourceCount: int64(len(source)),
		SourceRootSHA256: catalogFramedRoot(source), TargetCount: int64(len(target)),
		TargetRootSHA256: catalogFramedRoot(target)}
	canonical := fmt.Sprintf("catalog-v1-adoption-evidence.v1\x00%d\x00%s\x00%d\x00%s\x00%s\x00%s",
		evidence.SourceCount, evidence.SourceRootSHA256, evidence.TargetCount,
		evidence.TargetRootSHA256, evidence.BlockerSHA256, evidence.ConsumerSHA256)
	digest := sha256.Sum256([]byte(canonical))
	evidence.EvidenceSHA256 = hex.EncodeToString(digest[:])
	return evidence
}

func catalogFramedRoot(entries []catalogRootEntry) string {
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].kind+"\x00"+entries[left].identity <
			entries[right].kind+"\x00"+entries[right].identity
	})
	digest := sha256.New()
	for _, entry := range entries {
		catalogWriteFrame(digest, entry.kind)
		catalogWriteFrame(digest, entry.identity)
		catalogWriteFrame(digest, entry.document)
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func catalogWriteFrame(digest hash.Hash, value string) {
	_ = binary.Write(digest, binary.BigEndian, uint32(len([]byte(value))))
	_, _ = digest.Write([]byte(value))
}

func newCatalogCutoverBatch(authority catalogCutoverAuthority, checkpoint string,
	ordinal int, evidence catalogCutoverEvidence, mappings []catalogCutoverMapping,
) catalogCutoverBatch {
	changes := make([]any, 0, len(mappings))
	for index, mapping := range mappings {
		sourceDigest := sha256.Sum256([]byte(mapping.sourceDocument))
		key := map[string]any{"source_kind": checkpoint,
			"source_identity": mapping.sourceIdentity,
			"target_kind":     mapping.targetKind, "target_identity": mapping.targetIdentity}
		after := map[string]any{"source_kind": checkpoint,
			"source_identity": mapping.sourceIdentity,
			"source_sha256":   hex.EncodeToString(sourceDigest[:]),
			"source_document": json.RawMessage(mapping.sourceDocument),
			"target_kind":     mapping.targetKind, "target_identity": mapping.targetIdentity,
			"target_record": json.RawMessage(mapping.targetRecord)}
		changes = append(changes, map[string]any{"after": after, "before": nil,
			"key": key, "kind": "insert", "ordinal": index + 1})
	}
	sourceXID := fmt.Sprintf("%d", 6100+ordinal)
	document := map[string]any{"changes": changes, "checkpoint_key": checkpoint,
		"checkpoint_state": "complete", "last_key": evidence,
		"operation_id": authority.operationID, "relation_ordinal": 1,
		"run_id": authority.runID, "scope_id": authority.tenantID,
		"scope_kind": "tenant", "source_position": map[string]any{
			"batch_ordinal": ordinal, "commit_lsn": authority.finalLSN,
			"final_lsn": authority.finalLSN, "received_lsn": "0/A0",
			"snapshot_lsn": "0/20", "snapshot_name": "catalog-phase9-snapshot",
			"source_xid": sourceXID}}
	canonical, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(canonical)
	lastKey, err := json.Marshal(evidence)
	if err != nil {
		panic(err)
	}
	return catalogCutoverBatch{checkpoint: checkpoint, sourceXID: sourceXID,
		ordinal: ordinal, bytes: canonical, digest: hex.EncodeToString(digest[:]),
		lastKey: string(lastKey)}
}

func executeCatalogCutoverBatch(ctx context.Context, database *sql.DB,
	authority catalogCutoverAuthority, schema string, batch catalogCutoverBatch,
) (int64, string, string, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", "", err
	}
	rollback := func(cause error) (int64, string, string, error) {
		_ = transaction.Rollback()
		return 0, "", "", cause
	}
	if _, err = transaction.ExecContext(ctx,
		"SET LOCAL SESSION AUTHORIZATION harpy_adopter"); err != nil {
		return rollback(err)
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)", schema+",pg_catalog"); err != nil {
		return rollback(err)
	}
	contract, err := catalogadoption.NewCutoverContract()
	if err != nil {
		return rollback(err)
	}
	var count int64
	var keyDigest, stateDigest string
	err = transaction.QueryRowContext(ctx, contract.Query(), authority.operationID,
		authority.runID, "tenant", authority.tenantID, 1, batch.checkpoint,
		"catalog-phase9-snapshot", "0/20", "0/A0", authority.finalLSN,
		authority.finalLSN, batch.sourceXID, batch.ordinal, "complete", batch.lastKey,
		batch.bytes, batch.digest, int64(len(batch.bytes)), authority.leaseManifest,
	).Scan(&count, &keyDigest, &stateDigest)
	if err != nil {
		return rollback(err)
	}
	if err = transaction.Commit(); err != nil {
		return 0, "", "", err
	}
	return count, keyDigest, stateDigest, nil
}
