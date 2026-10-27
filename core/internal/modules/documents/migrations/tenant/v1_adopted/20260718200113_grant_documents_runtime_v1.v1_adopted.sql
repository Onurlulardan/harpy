-- +goose Up
GRANT SELECT ON
    documents_folders, documents_items, documents_versions, documents_uploads,
    documents_links, documents_legal_holds, documents_blob_deletions,
    documents_archives, documents_archive_entries, documents_operations,
    documents_kek_releases, documents_format_releases, documents_audit_entries,
    documents_compatibility_hits
TO harpy_runtime;

GRANT INSERT (id, parent_id, owner_id, display_name, lifecycle_state, row_version, created_at, updated_at)
ON documents_folders TO harpy_runtime;
GRANT UPDATE (parent_id, display_name, lifecycle_state, row_version, updated_at, trashed_at)
ON documents_folders TO harpy_runtime;
GRANT INSERT (id, folder_id, owner_id, current_version_id, lifecycle_state, row_version, created_at, updated_at)
ON documents_items TO harpy_runtime;
GRANT UPDATE (folder_id, current_version_id, lifecycle_state, resume_state, row_version, trashed_at, retention_ends_at, purge_requested_at, purge_starts_at, purge_reason, purged_at, tombstone_until, purge_evidence_sha256, updated_at)
ON documents_items TO harpy_runtime;
GRANT INSERT (id, item_id, version_number, lifecycle_state, original_name, declared_mime, detected_mime, file_kind, plaintext_bytes, ciphertext_bytes, plaintext_sha256, ciphertext_sha256, cipher_format, object_key, object_version, object_generation, key_id, wrapped_dek, scan_signature_revision, scanned_at, created_at)
ON documents_versions TO harpy_runtime;
GRANT UPDATE (lifecycle_state, delete_receipt_sha256, purged_at)
ON documents_versions TO harpy_runtime;
GRANT INSERT (id, actor_id, document_id, candidate_version_id, expected_current_version_id, folder_id, request_sha256, expected_content_sha256, expected_plaintext_bytes, ciphertext_upper_bound, original_name, declared_mime, billing_reservation_id, attempt_generation, lifecycle_state, cleanup_required, expires_at, created_at, updated_at)
ON documents_uploads TO harpy_runtime;
GRANT UPDATE (attempt_generation, lifecycle_state, lease_owner_id, lease_expires_at, cleanup_required, cleanup_receipt_sha256, stable_error_code, ready_status, ready_response, updated_at)
ON documents_uploads TO harpy_runtime;
GRANT INSERT (id, item_id, source_owner, source_kind, source_id, source_version_id, usage, reference_epoch, lifecycle_state, created_at)
ON documents_links TO harpy_runtime;
GRANT UPDATE (lifecycle_state, released_at) ON documents_links TO harpy_runtime;
GRANT INSERT (id, item_id, source_owner, source_case_id, evidence_sha256, lifecycle_state, placed_at)
ON documents_legal_holds TO harpy_runtime;
GRANT UPDATE (lifecycle_state, released_at) ON documents_legal_holds TO harpy_runtime;
GRANT INSERT (id, target_kind, target_id, object_key, object_version, object_generation, plaintext_bytes, ciphertext_bytes, lifecycle_state, lease_generation, created_at, updated_at)
ON documents_blob_deletions TO harpy_runtime;
GRANT UPDATE (lifecycle_state, lease_owner_id, lease_generation, lease_expires_at, provider_request_id, delete_response_sha256, absence_probe_sha256, billing_adjustment_receipt_id, confirmed_at, updated_at)
ON documents_blob_deletions TO harpy_runtime;
GRANT INSERT (id, actor_id, authorization_revision, job_id, retry_of_archive_id, request_sha256, billing_reservation_id, lifecycle_state, entry_count, plaintext_bytes, created_at, updated_at)
ON documents_archives TO harpy_runtime;
GRANT UPDATE (lifecycle_state, entry_count, entries_sha256, plaintext_bytes, ciphertext_bytes, artifact_sha256, object_key, object_version, object_generation, key_id, wrapped_dek, ready_at, expires_at, delete_receipt_sha256, purged_at, updated_at)
ON documents_archives TO harpy_runtime;
GRANT INSERT (archive_id, entry_order, document_id, version_id, plaintext_bytes, plaintext_sha256, added_at)
ON documents_archive_entries TO harpy_runtime;
GRANT INSERT (operation_id, operation_kind, actor_id, idempotency_key, request_sha256, aggregate_kind, aggregate_id, aggregate_version, response_status, response_document, occurred_at)
ON documents_operations TO harpy_runtime;
GRANT INSERT (key_id, lifecycle_state, key_material_sha256, reference_count, activated_at, retire_after, retired_at, retirement_evidence_sha256)
ON documents_kek_releases TO harpy_runtime;
GRANT UPDATE (lifecycle_state, reference_count, retire_after, retired_at, retirement_evidence_sha256)
ON documents_kek_releases TO harpy_runtime;
GRANT INSERT (format_id, lifecycle_state, decoder_sha256, reference_count, last_reference_purged_at, retire_after, retired_at, retirement_evidence_sha256)
ON documents_format_releases TO harpy_runtime;
GRANT UPDATE (lifecycle_state, reference_count, last_reference_purged_at, retire_after, retired_at, retirement_evidence_sha256)
ON documents_format_releases TO harpy_runtime;
GRANT INSERT (audit_id, operation_id, actor_id, action, aggregate_kind, aggregate_id, aggregate_version, reason, occurred_at)
ON documents_audit_entries TO harpy_runtime;
GRANT INSERT (route_contract, consumer_id, build_id, hit_count, first_hit_at, last_hit_at, retired_at)
ON documents_compatibility_hits TO harpy_runtime;
GRANT UPDATE (hit_count, last_hit_at, retired_at) ON documents_compatibility_hits TO harpy_runtime;

GRANT SELECT ON
    documents_folders, documents_items, documents_versions, documents_links,
    documents_legal_holds, documents_blob_deletions, documents_adoption_runs,
    documents_blob_copy_checkpoints, documents_adoption_issues,
    documents_compatibility_hits, documents_reverse_reconciliations
TO harpy_adopter;
GRANT INSERT (id, parent_id, owner_id, display_name, lifecycle_state, row_version, created_at, updated_at, trashed_at)
ON documents_folders TO harpy_adopter;
GRANT UPDATE (parent_id, display_name, lifecycle_state, row_version, updated_at, trashed_at)
ON documents_folders TO harpy_adopter;
GRANT INSERT (id, folder_id, owner_id, current_version_id, lifecycle_state, resume_state, row_version, trashed_at, retention_ends_at, purge_requested_at, purge_starts_at, purge_reason, purged_at, tombstone_until, purge_evidence_sha256, created_at, updated_at)
ON documents_items TO harpy_adopter;
GRANT UPDATE (folder_id, current_version_id, lifecycle_state, resume_state, row_version, trashed_at, retention_ends_at, purge_requested_at, purge_starts_at, purge_reason, purged_at, tombstone_until, purge_evidence_sha256, updated_at)
ON documents_items TO harpy_adopter;
GRANT INSERT (id, item_id, version_number, lifecycle_state, original_name, declared_mime, detected_mime, file_kind, plaintext_bytes, ciphertext_bytes, plaintext_sha256, ciphertext_sha256, cipher_format, object_key, object_version, object_generation, key_id, wrapped_dek, scan_signature_revision, scanned_at, delete_receipt_sha256, purged_at, created_at)
ON documents_versions TO harpy_adopter;
GRANT UPDATE (lifecycle_state, delete_receipt_sha256, purged_at) ON documents_versions TO harpy_adopter;
GRANT INSERT (id, item_id, source_owner, source_kind, source_id, source_version_id, usage, reference_epoch, lifecycle_state, created_at, released_at)
ON documents_links TO harpy_adopter;
GRANT UPDATE (lifecycle_state, released_at) ON documents_links TO harpy_adopter;
GRANT INSERT (id, item_id, source_owner, source_case_id, evidence_sha256, lifecycle_state, placed_at, released_at)
ON documents_legal_holds TO harpy_adopter;
GRANT UPDATE (lifecycle_state, released_at) ON documents_legal_holds TO harpy_adopter;
GRANT INSERT (id, target_kind, target_id, object_key, object_version, object_generation, plaintext_bytes, ciphertext_bytes, lifecycle_state, lease_owner_id, lease_generation, lease_expires_at, provider_request_id, delete_response_sha256, absence_probe_sha256, billing_adjustment_receipt_id, confirmed_at, created_at, updated_at)
ON documents_blob_deletions TO harpy_adopter;
GRANT UPDATE (lifecycle_state, lease_owner_id, lease_generation, lease_expires_at, provider_request_id, delete_response_sha256, absence_probe_sha256, billing_adjustment_receipt_id, confirmed_at, updated_at)
ON documents_blob_deletions TO harpy_adopter;
GRANT INSERT (run_id, cutover_generation, manifest_sha256, transform_version, lifecycle_state, source_count, target_count, source_plaintext_bytes, target_plaintext_bytes, source_merkle_sha256, target_merkle_sha256, pending_object_count, pending_job_count, final_lsn, applied_lsn, acceptance_sha256, created_at, updated_at)
ON documents_adoption_runs TO harpy_adopter;
GRANT UPDATE (lifecycle_state, source_count, target_count, source_plaintext_bytes, target_plaintext_bytes, source_merkle_sha256, target_merkle_sha256, pending_object_count, pending_job_count, final_lsn, applied_lsn, acceptance_sha256, updated_at)
ON documents_adoption_runs TO harpy_adopter;
GRANT INSERT (run_id, source_document_id, source_version_id, source_object_key, source_object_revision, source_bytes, source_sha256, target_document_id, target_version_id, target_object_key, target_object_version, target_ciphertext_bytes, target_plaintext_sha256, metadata_sha256, applied_lsn, lifecycle_state, issue_code, verified_at, source_hold_released_at, updated_at)
ON documents_blob_copy_checkpoints TO harpy_adopter;
GRANT UPDATE (target_object_key, target_object_version, target_ciphertext_bytes, target_plaintext_sha256, metadata_sha256, applied_lsn, lifecycle_state, issue_code, verified_at, source_hold_released_at, updated_at)
ON documents_blob_copy_checkpoints TO harpy_adopter;
GRANT INSERT (run_id, issue_id, source_version_id, issue_code, evidence_sha256, created_at)
ON documents_adoption_issues TO harpy_adopter;
GRANT INSERT (route_contract, consumer_id, build_id, hit_count, first_hit_at, last_hit_at, retired_at)
ON documents_compatibility_hits TO harpy_adopter;
GRANT UPDATE (hit_count, last_hit_at, retired_at) ON documents_compatibility_hits TO harpy_adopter;
GRANT INSERT (reconciliation_id, source_generation, target_generation, source_count, target_count, source_sha256, target_sha256, lifecycle_state, requested_by, approved_by, created_at, decided_at)
ON documents_reverse_reconciliations TO harpy_adopter;
GRANT UPDATE (lifecycle_state, approved_by, decided_at) ON documents_reverse_reconciliations TO harpy_adopter;

GRANT EXECUTE ON FUNCTION documents_guard_folder_cycle_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_folder_cycle_v1() TO harpy_adopter;
GRANT EXECUTE ON FUNCTION documents_guard_current_version_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_current_version_v1() TO harpy_adopter;
GRANT EXECUTE ON FUNCTION documents_guard_version_mutation_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_version_mutation_v1() TO harpy_adopter;
GRANT EXECUTE ON FUNCTION documents_guard_upload_transition_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_archive_entry_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_archive_entry_immutable_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_archive_transition_v1() TO harpy_runtime;
