-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_repair_page_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea
)
RETURNS TABLE (
    identifier_id uuid, user_id uuid, identifier_kind text,
    normalized_value text, normalized_sha256 bytea,
    identifier_state text, source_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.directory.repair', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM directory repair page authority is invalid';
    END IF;
    RETURN QUERY
    SELECT identifier.id, identifier.user_id, identifier.kind,
           identifier.normalized_value, identifier.normalized_sha256,
           identifier.state, identifier.revision
    FROM iam_user_identifiers AS identifier
    CROSS JOIN iam_maintenance_checkpoints AS checkpoint
    WHERE checkpoint.kind = 'iam.directory.repair'
      AND (checkpoint.cursor_id IS NULL OR identifier.id > checkpoint.cursor_id)
    ORDER BY identifier.id
    LIMIT 250;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_repair_page_v1(
    name,uuid,uuid,uuid,integer,text,bytea
) FROM PUBLIC;
