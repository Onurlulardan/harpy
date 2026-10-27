-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_repair_apply_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_identifier_id uuid, p_user_id uuid, p_kind text, p_state text,
    p_source_revision bigint, p_source_sha256 bytea, p_key_evidence jsonb
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_maintenance_claim_assert_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.directory.repair', p_payload_revision,
        p_payload_reference, p_payload_sha256
    ) OR p_identifier_id IS NULL OR p_user_id IS NULL OR
       p_kind NOT IN ('email_v1', 'portal_v1') OR
       p_state NOT IN ('pending', 'verified', 'active', 'tombstoned', 'retired') OR
       p_source_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (
           SELECT 1 FROM iam_user_identifiers AS identifier
           WHERE identifier.id = p_identifier_id AND identifier.user_id = p_user_id
             AND identifier.kind = p_kind AND identifier.state = p_state
             AND identifier.revision = p_source_revision
             AND identifier.normalized_sha256 = p_source_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM directory repair source is invalid';
    END IF;
    RETURN p_source_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_repair_apply_v1(
    name,uuid,uuid,uuid,integer,text,bytea,
    uuid,uuid,text,text,bigint,bytea,jsonb
) FROM PUBLIC;
