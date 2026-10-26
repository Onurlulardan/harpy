//go:build integration

package integrationtest

import (
	"encoding/json"
	"testing"
)

const documentsCutoverObservedAt = "2026-07-21T18:00:00+00:00"

func documentsCutoverChanges(checkpointKey, folderName string) []any {
	switch checkpointKey {
	case "inventory":
		return []any{documentsCutoverChange("inventory", "inventory:legacy", "inventory",
			documentsAdoptionRun, map[string]any{
				"adoption_run_id":    documentsAdoptionRun,
				"created_at":         documentsCutoverObservedAt,
				"cutover_generation": 13,
				"manifest_sha256":    documentsCutoverManifest,
				"transform_version":  "documents_adoption_v1",
			})}
	case "folder":
		return []any{documentsCutoverChange("folder", "folder:legacy", "folder",
			documentsTargetFolder, map[string]any{
				"id": documentsTargetFolder, "parent_id": nil,
				"owner_id":     "018f0000-0000-7000-8000-000000000030",
				"display_name": folderName, "lifecycle_state": "ACTIVE",
				"row_version": 1, "created_at": documentsCutoverObservedAt,
				"updated_at": documentsCutoverObservedAt, "trashed_at": nil,
			})}
	case "item":
		return []any{documentsCutoverChange("item", "item:legacy", "item",
			documentsTargetItem, map[string]any{
				"id": documentsTargetItem, "folder_id": documentsTargetFolder,
				"owner_id":           "018f0000-0000-7000-8000-000000000030",
				"current_version_id": documentsTargetVersion,
				"lifecycle_state":    "ACTIVE", "resume_state": nil,
				"reference_epoch": 1, "row_version": 1, "description": "",
				"tags": []any{}, "trashed_at": nil, "retention_ends_at": nil,
				"purge_requested_at": nil, "purge_starts_at": nil,
				"purge_reason": nil, "purged_at": nil, "tombstone_until": nil,
				"purge_evidence_sha256": nil, "created_at": documentsCutoverObservedAt,
				"updated_at": documentsCutoverObservedAt,
			})}
	case "version":
		version := documentsCutoverChange("version", "version:legacy", "version",
			documentsTargetVersion, map[string]any{
				"id": documentsTargetVersion, "item_id": documentsTargetItem,
				"version_number": 1, "lifecycle_state": "READY",
				"original_name": "evidence.pdf", "declared_mime": "application/pdf",
				"detected_mime": "application/pdf", "file_kind": "PDF",
				"plaintext_bytes": 11, "plaintext_sha256": "\\x" + documentsSourceContentHash,
				"scan_signature_revision": "clamav-1", "scanned_at": documentsCutoverObservedAt,
				"created_at": documentsCutoverObservedAt,
			})
		checkpoint := documentsCutoverChange("version", "version:legacy", "checkpoint",
			documentsTargetVersion, map[string]any{
				"adoption_run_id": documentsAdoptionRun,
				"enqueued_at":     documentsCutoverObservedAt, "job_id": documentsCopyJob,
				"source_bytes": 11, "source_document_id": documentsSourceDocument,
				"source_lsn": "0/20", "source_metadata_sha256": documentsSourceMetaHash,
				"source_object_key":      "legacy/evidence.pdf",
				"source_object_revision": "legacy-version-1",
				"source_sha256":          documentsSourceContentHash,
				"source_version_id":      documentsSourceVersion,
				"target_document_id":     documentsTargetItem,
				"target_version_id":      documentsTargetVersion,
			})
		return []any{version, checkpoint}
	default:
		return nil
	}
}

func documentsCutoverChange(sourceKind, sourceIdentity, targetKind,
	targetID string, record map[string]any,
) map[string]any {
	targetIdentity := targetKind + ":" + targetID
	value := map[string]any{
		"source_identity": sourceIdentity, "source_kind": sourceKind,
		"source_sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"target_id":     targetID, "target_identity": targetIdentity,
		"target_kind": targetKind, "target_record": record,
	}
	return map[string]any{
		"after": value, "before": nil,
		"key": map[string]any{
			"source_identity": sourceIdentity, "source_kind": sourceKind,
			"target_identity": targetIdentity, "target_kind": targetKind,
		},
		"kind": "insert", "ordinal": 1,
	}
}

func documentsCutoverBatch(t *testing.T, checkpointKey, state string,
	batchOrdinal int, sourceXID string, changes []any,
) []byte {
	t.Helper()
	for index := range changes {
		changes[index].(map[string]any)["ordinal"] = index + 1
	}
	finalLSN := any(nil)
	lastKey := any(map[string]any{"source": checkpointKey})
	commitLSN := "0/20"
	if state == "complete" {
		finalLSN, lastKey, commitLSN = "0/30", nil, "0/30"
	}
	document := map[string]any{
		"changes": changes, "checkpoint_key": checkpointKey,
		"checkpoint_state": state, "last_key": lastKey,
		"operation_id": documentsCutoverOperation, "relation_ordinal": 1,
		"run_id": documentsCutoverRun, "scope_id": documentsCutoverTenantID,
		"scope_kind": "tenant", "source_position": map[string]any{
			"batch_ordinal": batchOrdinal, "commit_lsn": commitLSN,
			"final_lsn": finalLSN, "received_lsn": commitLSN,
			"snapshot_lsn": "0/10", "snapshot_name": "documents_snapshot",
			"source_xid": sourceXID,
		},
	}
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
