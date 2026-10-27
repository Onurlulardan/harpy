-- +goose Up
REVOKE SELECT ON
    notifications_inbox_counters,notifications_messages,
    notifications_effect_receipts,notifications_deliveries,
    notifications_delivery_attempts,notifications_delivery_receipts,
    notifications_suppressed_targets,notifications_preferences,
    notifications_templates,notifications_template_versions,
    notifications_provider_secrets,notifications_provider_configs,
    notifications_provider_config_versions,notifications_audit_entries,
    notifications_realtime_hints,notifications_operator_alerts,
    notifications_writer_authority,notifications_adoption_runs,
    notifications_adoption_checkpoints,notifications_adoption_issues,
    notifications_migration_intents,notifications_migration_receipts,
    notifications_compatibility_hits,notifications_cutover_evidence,
    notifications_retention_checkpoints,notifications_retention_receipts,
    notifications_v1_adoption_stage,notifications_v1_cutover_receipts
FROM harpy_adopter;

REVOKE INSERT (id,recipient_id,template_version_id,producer,notification_type,category,template_kind,locale,subject,body,data_digest,subject_digest,body_digest,row_version,inbox_revision,created_at,read_at,dismissed_at,purged_at) ON notifications_messages FROM harpy_adopter;
REVOKE UPDATE (subject,body,row_version,read_at,dismissed_at,purged_at) ON notifications_messages FROM harpy_adopter;
REVOKE INSERT (producer,effect_key_hash,recipient_id,canonical_digest,message_id,legacy_receipt_reference,outcome,created_at,retain_until) ON notifications_effect_receipts FROM harpy_adopter;
REVOKE UPDATE (outcome) ON notifications_effect_receipts FROM harpy_adopter;
REVOKE INSERT (id,message_id,provider_config_version_id,channel,state,attempt,maximum_attempts,lease_owner,lease_generation,lease_until,not_before,outcome_code,outcome_unknown,mandatory,row_version) ON notifications_deliveries FROM harpy_adopter;
REVOKE UPDATE (state,attempt,lease_owner,lease_generation,lease_until,not_before,outcome_code,outcome_unknown,row_version) ON notifications_deliveries FROM harpy_adopter;
REVOKE INSERT (recipient_id,category,channel,enabled,explicit,verified_target,quiet_timezone,quiet_start_minute,quiet_end_minute,row_version,updated_at) ON notifications_preferences FROM harpy_adopter;
REVOKE UPDATE (enabled,explicit,verified_target,quiet_timezone,quiet_start_minute,quiet_end_minute,row_version,updated_at) ON notifications_preferences FROM harpy_adopter;
REVOKE INSERT (id,notification_type,published_version_id,row_version,created_at,updated_at) ON notifications_templates FROM harpy_adopter;
REVOKE UPDATE (published_version_id,row_version,updated_at) ON notifications_templates FROM harpy_adopter;
REVOKE INSERT (id,ciphertext,nonce,wrapped_dek,kek_revision,plaintext_digest,created_at,rewrapped_at) ON notifications_provider_secrets FROM harpy_adopter;
REVOKE UPDATE (wrapped_dek,kek_revision,rewrapped_at) ON notifications_provider_secrets FROM harpy_adopter;
REVOKE INSERT (id,provider_kind,name,enabled,current_version_id,row_version,created_at,updated_at) ON notifications_provider_configs FROM harpy_adopter;
REVOKE UPDATE (name,enabled,current_version_id,row_version,updated_at) ON notifications_provider_configs FROM harpy_adopter;
REVOKE INSERT (id,source_generation,source_manifest_digest,lifecycle_state,started_at,completed_at,row_version) ON notifications_adoption_runs FROM harpy_adopter;
REVOKE UPDATE (lifecycle_state,completed_at,row_version) ON notifications_adoption_runs FROM harpy_adopter;
REVOKE INSERT (run_id,source_table,last_source_id,adopted_count,source_checksum,target_checksum,updated_at) ON notifications_adoption_checkpoints FROM harpy_adopter;
REVOKE UPDATE (last_source_id,adopted_count,source_checksum,target_checksum,updated_at) ON notifications_adoption_checkpoints FROM harpy_adopter;
REVOKE INSERT (id,run_id,source_table,source_key_digest,code,created_at,resolved_at) ON notifications_adoption_issues FROM harpy_adopter;
REVOKE UPDATE (resolved_at) ON notifications_adoption_issues FROM harpy_adopter;
REVOKE INSERT (singleton,writer,lifecycle_state,generation,row_version,updated_at) ON notifications_writer_authority FROM harpy_adopter;
REVOKE UPDATE (writer,lifecycle_state,generation,row_version,updated_at) ON notifications_writer_authority FROM harpy_adopter;
REVOKE INSERT (producer,effect_key_hash,recipient_id,producer_generation,notification_generation,canonical_digest,payload_digest,payload,job_id,state,created_at,updated_at,row_version) ON notifications_migration_intents FROM harpy_adopter;
REVOKE UPDATE (state,updated_at,row_version) ON notifications_migration_intents FROM harpy_adopter;
REVOKE INSERT (producer,effect_key_hash,recipient_id,producer_generation,notification_generation,v1_receipt_digest,terminal_outcome,created_at) ON notifications_migration_receipts FROM harpy_adopter;
REVOKE UPDATE (terminal_outcome) ON notifications_migration_receipts FROM harpy_adopter;
REVOKE INSERT (id,run_id,source_queue_depth,future_delivery_count,running_delivery_count,unresolved_provider_receipt_count,legacy_connection_count,source_evidence_digest,observed_at,expires_at) ON notifications_cutover_evidence FROM harpy_adopter;
REVOKE UPDATE (source_queue_depth,future_delivery_count,running_delivery_count,unresolved_provider_receipt_count,legacy_connection_count,source_evidence_digest,observed_at,expires_at) ON notifications_cutover_evidence FROM harpy_adopter;
REVOKE INSERT (id,delivery_id,attempt_number,lease_generation,provider_kind,response_class,provider_reference_digest,status_code,error_code,started_at,finished_at) ON notifications_delivery_attempts FROM harpy_adopter;
REVOKE INSERT (id,delivery_id,provider_reference_digest,outcome,signature_key_id,received_at) ON notifications_delivery_receipts FROM harpy_adopter;
REVOKE INSERT (id,template_id,version_number,locale,template_kind,subject_pattern,body_pattern,variables,mandatory_email,content_digest,published_at) ON notifications_template_versions FROM harpy_adopter;
REVOKE INSERT (id,config_id,version_number,secret_id,metadata,metadata_digest,created_at) ON notifications_provider_config_versions FROM harpy_adopter;
REVOKE INSERT (id,action,resource_id,actor_id,resource_revision,occurred_at) ON notifications_audit_entries FROM harpy_adopter;
REVOKE INSERT (id,contract_name,consumer_manifest_digest,occurred_at) ON notifications_compatibility_hits FROM harpy_adopter;

REVOKE SELECT ON notifications_migration_intents,notifications_migration_receipts FROM harpy_runtime;
REVOKE INSERT (producer,effect_key_hash,recipient_id,producer_generation,notification_generation,canonical_digest,payload_digest,payload,job_id,state,created_at,updated_at,row_version) ON notifications_migration_intents FROM harpy_runtime;
REVOKE UPDATE (state,updated_at,row_version) ON notifications_migration_intents FROM harpy_runtime;
REVOKE INSERT (producer,effect_key_hash,recipient_id,producer_generation,notification_generation,v1_receipt_digest,terminal_outcome,created_at) ON notifications_migration_receipts FROM harpy_runtime;
REVOKE UPDATE (terminal_outcome) ON notifications_migration_receipts FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION notifications_retention_redact_content_v1(
    timestamptz,integer
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_retention_purge_attempts_v1(
    timestamptz,integer
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_retention_purge_receipts_v1(
    timestamptz,integer
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_retention_purge_effects_v1(
    timestamptz,integer
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_retention_purge_hints_v1(
    timestamptz,integer
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_retention_purge_deliveries_v1(
    timestamptz,integer
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_retention_purge_messages_v1(
    timestamptz,integer
) FROM harpy_adopter;

REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_target_record_valid_v1(
    text,text,uuid,jsonb
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_value_valid_v1(
    jsonb,jsonb
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_batch_decode_v1(
    bytea,text,bigint
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_batch_assert_v1(
    jsonb,uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,
    pg_lsn,pg_lsn,xid8,integer,text,jsonb
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,
    xid8,integer,jsonb,text
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_project_message_v1(jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_project_effect_v1(jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_project_delivery_v1(jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_project_attempt_v1(jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_project_receipt_v1(jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_queue_evidence_v1(
    uuid,uuid,uuid
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_provider_evidence_v1(uuid)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_legacy_evidence_v1(
    uuid,uuid,uuid
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_bridge_disabled_v1(uuid)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_projection_ready_v1(
    uuid,uuid,uuid
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION notifications_v1_adoption_project_v1(
    uuid,uuid,uuid
) FROM harpy_adopter;

-- +goose StatementBegin
CREATE FUNCTION notifications_v1_cutover_receipt_reject_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Notifications V1 cutover receipt is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_cutover_receipt_reject_mutation_v1()
FROM PUBLIC;
CREATE TRIGGER notifications_v1_cutover_receipt_immutable
BEFORE UPDATE OR DELETE ON notifications_v1_cutover_receipts
FOR EACH ROW EXECUTE FUNCTION notifications_v1_cutover_receipt_reject_mutation_v1();
ALTER TABLE notifications_v1_cutover_receipts
    ENABLE ALWAYS TRIGGER notifications_v1_cutover_receipt_immutable;


