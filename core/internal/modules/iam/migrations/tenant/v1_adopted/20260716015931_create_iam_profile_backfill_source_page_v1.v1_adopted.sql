-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_backfill_source_page_v1(
    p_resource_kind text, p_source_schema_id uuid,
    p_after_resource_id uuid, p_through_resource_id uuid
)
RETURNS TABLE (source_count integer, source_sha256 bytea)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_id uuid;
    v_revision bigint;
    v_profile_sha256 bytea;
    v_part bytea;
    v_evidence bytea;
    v_count integer;
BEGIN
    v_evidence := ''::bytea;
    v_count := 0;
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_source_schema_id IS NULL OR p_through_resource_id IS NULL OR
       (p_after_resource_id IS NOT NULL AND
        p_after_resource_id >= p_through_resource_id) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill source page input is invalid';
    END IF;
    FOR v_id,v_revision,v_profile_sha256 IN
        SELECT resource.id, resource.revision, resource.profile_sha256
        FROM (
            SELECT usr.id, usr.revision, usr.profile_sha256
            FROM iam_users AS usr
            WHERE p_resource_kind = 'user'
              AND usr.profile_schema_version_id = p_source_schema_id
            UNION ALL
            SELECT org.id, org.revision, org.profile_sha256
            FROM iam_organizations AS org
            WHERE p_resource_kind = 'organization'
              AND org.profile_schema_version_id = p_source_schema_id
            UNION ALL
            SELECT role.id, role.revision, role.profile_sha256
            FROM iam_roles AS role
            WHERE p_resource_kind = 'role'
              AND role.profile_schema_version_id = p_source_schema_id
        ) AS resource
        WHERE (p_after_resource_id IS NULL OR resource.id > p_after_resource_id)
          AND resource.id <= p_through_resource_id
        ORDER BY resource.id
    LOOP
        v_part := pg_catalog.convert_to(v_id::text, 'UTF8');
        v_evidence := v_evidence || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_part := pg_catalog.convert_to(v_revision::text, 'UTF8');
        v_evidence := v_evidence || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_part := v_profile_sha256;
        v_evidence := v_evidence || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_count := v_count + 1;
    END LOOP;
    RETURN QUERY SELECT v_count, pg_catalog.sha256(v_evidence);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_backfill_source_page_v1(
    text,uuid,uuid,uuid
) FROM PUBLIC;
