-- +goose Up
REVOKE SELECT ON
    documents_folders,documents_items,documents_versions,documents_links,
    documents_legal_holds,documents_blob_deletions,documents_adoption_runs,
    documents_blob_copy_checkpoints,documents_adoption_issues,
    documents_compatibility_hits,documents_reverse_reconciliations,
    documents_upload_attempts
FROM forgecrud_adopter;
REVOKE INSERT (id,parent_id,owner_id,display_name,lifecycle_state,row_version,created_at,updated_at,trashed_at)
ON documents_folders FROM forgecrud_adopter;
REVOKE UPDATE (parent_id,display_name,lifecycle_state,row_version,updated_at,trashed_at)
ON documents_folders FROM forgecrud_adopter;
REVOKE INSERT (id,folder_id,owner_id,current_version_id,lifecycle_state,resume_state,reference_epoch,row_version,description,tags,trashed_at,retention_ends_at,purge_requested_at,purge_starts_at,purge_reason,purged_at,tombstone_until,purge_evidence_sha256,created_at,updated_at)
ON documents_items FROM forgecrud_adopter;
REVOKE UPDATE (folder_id,current_version_id,lifecycle_state,resume_state,reference_epoch,row_version,description,tags,trashed_at,retention_ends_at,purge_requested_at,purge_starts_at,purge_reason,purged_at,tombstone_until,purge_evidence_sha256,updated_at)
ON documents_items FROM forgecrud_adopter;
REVOKE INSERT (id,item_id,version_number,lifecycle_state,original_name,declared_mime,detected_mime,file_kind,plaintext_bytes,ciphertext_bytes,plaintext_sha256,ciphertext_sha256,cipher_format,object_key,object_version,object_generation,key_id,wrapped_dek,scan_signature_revision,scanned_at,delete_receipt_sha256,purged_at,created_at)
ON documents_versions FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,delete_receipt_sha256,purged_at)
ON documents_versions FROM forgecrud_adopter;
REVOKE INSERT (id,item_id,source_owner,source_kind,source_id,source_version_id,usage,reference_epoch,lifecycle_state,created_at,released_at)
ON documents_links FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,released_at) ON documents_links FROM forgecrud_adopter;
REVOKE INSERT (id,item_id,source_owner,source_case_id,evidence_sha256,lifecycle_state,placed_at,released_at)
ON documents_legal_holds FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,released_at)
ON documents_legal_holds FROM forgecrud_adopter;
REVOKE INSERT (id,target_kind,target_id,object_key,object_version,object_generation,plaintext_bytes,ciphertext_bytes,lifecycle_state,lease_owner_id,lease_generation,lease_expires_at,provider_request_id,delete_response_sha256,absence_probe_sha256,billing_adjustment_receipt_id,confirmed_at,created_at,updated_at)
ON documents_blob_deletions FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,lease_owner_id,lease_generation,lease_expires_at,provider_request_id,delete_response_sha256,absence_probe_sha256,billing_adjustment_receipt_id,confirmed_at,updated_at)
ON documents_blob_deletions FROM forgecrud_adopter;
REVOKE INSERT (run_id,cutover_generation,manifest_sha256,transform_version,lifecycle_state,source_count,target_count,source_plaintext_bytes,target_plaintext_bytes,source_merkle_sha256,target_merkle_sha256,pending_object_count,pending_job_count,final_lsn,applied_lsn,acceptance_sha256,created_at,updated_at)
ON documents_adoption_runs FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,source_count,target_count,source_plaintext_bytes,target_plaintext_bytes,source_merkle_sha256,target_merkle_sha256,pending_object_count,pending_job_count,final_lsn,applied_lsn,acceptance_sha256,updated_at)
ON documents_adoption_runs FROM forgecrud_adopter;
REVOKE INSERT (run_id,source_document_id,source_version_id,source_object_key,source_object_revision,source_bytes,source_sha256,source_metadata_sha256,source_lsn,target_document_id,target_version_id,target_object_key,target_object_version,target_object_generation,target_ciphertext_bytes,target_plaintext_sha256,target_ciphertext_sha256,target_key_id,target_wrapped_dek,metadata_sha256,applied_lsn,lifecycle_state,issue_code,verified_at,source_hold_released_at,updated_at)
ON documents_blob_copy_checkpoints FROM forgecrud_adopter;
REVOKE UPDATE (source_metadata_sha256,source_lsn,target_object_key,target_object_version,target_object_generation,target_ciphertext_bytes,target_plaintext_sha256,target_ciphertext_sha256,target_key_id,target_wrapped_dek,metadata_sha256,applied_lsn,lifecycle_state,issue_code,verified_at,source_hold_released_at,updated_at)
ON documents_blob_copy_checkpoints FROM forgecrud_adopter;
REVOKE INSERT (run_id,issue_id,source_version_id,issue_code,evidence_sha256,created_at)
ON documents_adoption_issues FROM forgecrud_adopter;
REVOKE INSERT (route_contract,consumer_id,build_id,hit_count,first_hit_at,last_hit_at,retired_at)
ON documents_compatibility_hits FROM forgecrud_adopter;
REVOKE UPDATE (hit_count,last_hit_at,retired_at)
ON documents_compatibility_hits FROM forgecrud_adopter;
REVOKE INSERT (reconciliation_id,source_generation,target_generation,source_count,target_count,source_sha256,target_sha256,lifecycle_state,requested_by,approved_by,created_at,decided_at)
ON documents_reverse_reconciliations FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,approved_by,decided_at)
ON documents_reverse_reconciliations FROM forgecrud_adopter;
REVOKE INSERT (upload_id,attempt_generation,effect_receipt_id,lifecycle_state,object_key,object_version,object_generation,plaintext_bytes,ciphertext_bytes,plaintext_sha256,ciphertext_sha256,detected_mime,file_kind,key_id,wrapped_dek,scan_signature_revision,scanned_at,cleanup_deletion_id,cleanup_receipt_sha256,created_at,updated_at)
ON documents_upload_attempts FROM forgecrud_adopter;
REVOKE UPDATE (lifecycle_state,object_key,object_version,object_generation,plaintext_bytes,ciphertext_bytes,plaintext_sha256,ciphertext_sha256,detected_mime,file_kind,key_id,wrapped_dek,scan_signature_revision,scanned_at,cleanup_deletion_id,cleanup_receipt_sha256,updated_at)
ON documents_upload_attempts FROM forgecrud_adopter;
REVOKE INSERT (cancel_requested_at) ON documents_uploads FROM forgecrud_adopter;
REVOKE UPDATE (cancel_requested_at) ON documents_uploads FROM forgecrud_adopter;

REVOKE EXECUTE ON FUNCTION documents_guard_folder_cycle_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_current_version_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_version_mutation_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_upload_transition_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_archive_entry_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_archive_entry_immutable_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_archive_transition_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_reference_target_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_guard_reference_transition_v1()
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_item_response_v1(uuid)
FROM forgecrud_adopter;

REVOKE EXECUTE ON FUNCTION documents_v1_adoption_target_record_valid_v1(
    text,text,uuid,jsonb
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_value_valid_v1(jsonb,jsonb)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_batch_decode_v1(bytea,text,bigint)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_batch_assert_v1(
    jsonb,uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,
    pg_lsn,pg_lsn,xid8,integer,text,jsonb
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,
    xid8,integer,jsonb,text
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_project_folder_v1(jsonb)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_project_item_v1(jsonb)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_project_version_v1(uuid,jsonb)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_project_inventory_v1(
    uuid,uuid,uuid,jsonb
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_project_checkpoint_v1(
    uuid,uuid,uuid,jsonb
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_mapping_ready_v1(uuid,uuid)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_projection_ready_v1(
    uuid,uuid,uuid
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_project_rows_v1(uuid,uuid,uuid)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_adoption_accept_v1(uuid,uuid,uuid)
FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_reverse_evidence_v1(
    uuid,uuid,bigint,bigint,bigint,bigint,bytea,bytea,uuid,uuid,timestamptz
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION documents_v1_reverse_accept_v1(
    uuid,uuid,bigint,bigint,bigint,bigint,bytea,bytea,uuid,uuid,timestamptz
) FROM forgecrud_adopter;

GRANT SELECT ON documents_adoption_runs,documents_blob_copy_checkpoints
TO forgecrud_runtime;
GRANT UPDATE (lifecycle_state,updated_at)
ON documents_adoption_runs TO forgecrud_runtime;
GRANT UPDATE (
    target_object_key,target_object_version,target_object_generation,
    target_ciphertext_bytes,target_plaintext_sha256,target_ciphertext_sha256,
    target_key_id,target_wrapped_dek,metadata_sha256,applied_lsn,
    lifecycle_state,issue_code,verified_at,updated_at
) ON documents_blob_copy_checkpoints TO forgecrud_runtime;

-- +goose StatementBegin
CREATE FUNCTION documents_v1_receipt_reject_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Documents V1 receipt is immutable';
END;
$function$;

REVOKE ALL ON FUNCTION documents_v1_receipt_reject_mutation_v1() FROM PUBLIC;
CREATE TRIGGER documents_v1_cutover_receipt_immutable
BEFORE UPDATE OR DELETE ON documents_v1_cutover_receipts
FOR EACH ROW EXECUTE FUNCTION documents_v1_receipt_reject_mutation_v1();
ALTER TABLE documents_v1_cutover_receipts
    ENABLE ALWAYS TRIGGER documents_v1_cutover_receipt_immutable;
CREATE TRIGGER documents_v1_reverse_receipt_immutable
BEFORE UPDATE OR DELETE ON documents_v1_reverse_receipts
FOR EACH ROW EXECUTE FUNCTION documents_v1_receipt_reject_mutation_v1();
ALTER TABLE documents_v1_reverse_receipts
    ENABLE ALWAYS TRIGGER documents_v1_reverse_receipt_immutable;

CREATE FUNCTION documents_v1_reverse_reconciliation_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP='INSERT' AND SESSION_USER='forgecrud_cutover_finalizer'
       AND CURRENT_USER='forgecrud_owner' AND NEW.lifecycle_state='ACCEPTED'
       AND NEW.approved_by IS NOT NULL AND NEW.decided_at IS NOT NULL THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Documents V1 reverse reconciliation authority denied';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_reverse_reconciliation_guard_v1() FROM PUBLIC;
CREATE TRIGGER documents_v1_reverse_reconciliation_guard
BEFORE INSERT OR UPDATE OR DELETE ON documents_reverse_reconciliations
FOR EACH ROW EXECUTE FUNCTION documents_v1_reverse_reconciliation_guard_v1();
ALTER TABLE documents_reverse_reconciliations
    ENABLE ALWAYS TRIGGER documents_v1_reverse_reconciliation_guard;
