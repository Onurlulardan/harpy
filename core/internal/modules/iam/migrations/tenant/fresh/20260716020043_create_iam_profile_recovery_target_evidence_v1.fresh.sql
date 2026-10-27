-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_target_evidence_v1(
    p_resource_kind text, p_target_schema_id uuid
)
RETURNS TABLE (target_count bigint, target_sha256 bytea)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_id uuid;
    v_schema_id uuid;
    v_revision bigint;
    v_profile_sha256 bytea;
    v_part bytea;
    v_entry bytea;
    v_evidence bytea;
    v_count bigint;
BEGIN
    v_evidence := pg_catalog.sha256(''::bytea);
    v_count := 0;
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_resource_kind NOT IN ('user','organization','role') OR
       p_target_schema_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile recovery target evidence input is invalid';
    END IF;
    FOR v_id,v_schema_id,v_revision,v_profile_sha256 IN
        SELECT resource.id, resource.profile_schema_version_id,
               resource.revision, resource.profile_sha256
        FROM (
            SELECT usr.id, usr.profile_schema_version_id,
                   usr.revision, usr.profile_sha256
            FROM iam_users AS usr WHERE p_resource_kind = 'user'
            UNION ALL
            SELECT org.id, org.profile_schema_version_id,
                   org.revision, org.profile_sha256
            FROM iam_organizations AS org WHERE p_resource_kind = 'organization'
            UNION ALL
            SELECT role.id, role.profile_schema_version_id,
                   role.revision, role.profile_sha256
            FROM iam_roles AS role WHERE p_resource_kind = 'role'
        ) AS resource
        WHERE resource.profile_schema_version_id = p_target_schema_id
        ORDER BY resource.id
    LOOP
        v_entry := ''::bytea;
        v_part := pg_catalog.convert_to(v_id::text, 'UTF8');
        v_entry := v_entry || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_part := pg_catalog.convert_to(v_schema_id::text, 'UTF8');
        v_entry := v_entry || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_part := pg_catalog.convert_to(v_revision::text, 'UTF8');
        v_entry := v_entry || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_part := v_profile_sha256;
        v_entry := v_entry || pg_catalog.int8send(
            pg_catalog.octet_length(v_part)::bigint
        ) || v_part;
        v_evidence := pg_catalog.sha256(
            v_evidence || pg_catalog.sha256(v_entry)
        );
        v_count := v_count + 1;
    END LOOP;
    RETURN QUERY SELECT v_count, v_evidence;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_target_evidence_v1(text,uuid)
    FROM PUBLIC;
