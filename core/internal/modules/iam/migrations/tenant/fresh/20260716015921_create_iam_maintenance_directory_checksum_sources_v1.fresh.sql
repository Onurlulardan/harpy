-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_checksum_sources_v1(
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
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.directory.checksum', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM directory checksum source authority is invalid';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM iam_user_identifiers) > 10000 THEN
        RAISE EXCEPTION USING ERRCODE = '54000',
            MESSAGE = 'IAM directory checksum source bound exceeded';
    END IF;
    RETURN QUERY
    SELECT identifier.id, identifier.user_id, identifier.kind,
           identifier.normalized_value, identifier.normalized_sha256,
           identifier.state, identifier.revision
    FROM iam_user_identifiers AS identifier
    ORDER BY identifier.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_checksum_sources_v1(
    name,uuid,uuid,uuid,integer,text,bytea
) FROM PUBLIC;
