-- +goose Up
REVOKE ALL ON TABLE audit_entries,audit_v1_adoption_runs,
    audit_v1_adoption_mappings,audit_v1_adoption_receipts
FROM PUBLIC,harpy_runtime,harpy_adopter;
REVOKE ALL ON FUNCTION audit_v1_derived_id_v1(uuid,text),
    audit_v1_redacted_changes_v1(jsonb,jsonb,jsonb),
    audit_v1_owner_module_v1(text),audit_v1_operation_v1(text,text),
    audit_v1_run_document_v1(audit_v1_adoption_runs),
    audit_v1_adopt_batch_v1(uuid,uuid,integer,timestamptz)
FROM PUBLIC,harpy_runtime,harpy_adopter;
GRANT EXECUTE ON FUNCTION audit_v1_adopt_batch_v1(uuid,uuid,integer,timestamptz)
TO harpy_adopter;
