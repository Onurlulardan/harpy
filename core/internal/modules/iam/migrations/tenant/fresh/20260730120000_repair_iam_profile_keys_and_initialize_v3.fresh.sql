-- +goose Up
-- Migration stream version: 20260730120000.
-- +goose StatementBegin
CREATE FUNCTION iam_profile_installation_legacy_document_v1(
    p_profile_values jsonb, p_profile_sha256 bytea,
    p_stable_field_id uuid, p_field_key text,
    p_value_kind text, p_cardinality text, p_field_state text
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_profile_values IS NULL OR p_stable_field_id IS NULL
       OR p_field_key IS DISTINCT FROM 'display_name'
       OR p_value_kind IS DISTINCT FROM 'string'
       OR p_cardinality IS DISTINCT FROM 'single'
       OR p_field_state IS DISTINCT FROM 'active'
       OR pg_catalog.jsonb_typeof(p_profile_values) <> 'object'
       OR (SELECT pg_catalog.count(*)
           FROM pg_catalog.jsonb_object_keys(p_profile_values)) <> 1
       OR NOT p_profile_values ? 'display_name'
       OR p_profile_values ? p_stable_field_id::text
       OR pg_catalog.jsonb_typeof(
            p_profile_values -> 'display_name'
          ) <> 'string'
       OR p_profile_sha256 IS DISTINCT FROM pg_catalog.sha256(
            pg_catalog.convert_to(p_profile_values::text, 'UTF8')
          ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM installation legacy profile shape is invalid';
    END IF;
    RETURN '{' || pg_catalog.to_json(p_stable_field_id::text)::text ||
        ':' || pg_catalog.to_json(
            p_profile_values ->> 'display_name'
        )::text || '}';
END;
$function$;
CREATE FUNCTION iam_profile_display_name_key_repair_v1(p_legacy_key text)
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_changed bigint;
    v_count bigint;
BEGIN
    v_changed := 0;
    IF p_legacy_key IS DISTINCT FROM 'display_name' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM legacy profile key is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT subject.id, pg_catalog.count(snapshot.field_id) AS field_count
            FROM iam_users AS subject LEFT JOIN iam_profile_field_revisions AS snapshot
              ON snapshot.schema_version_id = subject.profile_schema_version_id
             AND snapshot.resource_kind = 'user'
             AND snapshot.field_key = 'display_name'
            WHERE subject.profile_values ? 'display_name' GROUP BY subject.id
            UNION ALL
            SELECT subject.id, pg_catalog.count(snapshot.field_id)
            FROM iam_organizations AS subject LEFT JOIN iam_profile_field_revisions AS snapshot
              ON snapshot.schema_version_id = subject.profile_schema_version_id
             AND snapshot.resource_kind = 'organization'
             AND snapshot.field_key = 'display_name'
            WHERE subject.profile_values ? 'display_name' GROUP BY subject.id
            UNION ALL
            SELECT subject.id, pg_catalog.count(snapshot.field_id)
            FROM iam_roles AS subject LEFT JOIN iam_profile_field_revisions AS snapshot
              ON snapshot.schema_version_id = subject.profile_schema_version_id
             AND snapshot.resource_kind = 'role'
             AND snapshot.field_key = 'display_name'
            WHERE subject.profile_values ? 'display_name' GROUP BY subject.id
        ) AS installation_profile(id, field_count)
        WHERE installation_profile.field_count <> 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM installation profile snapshot is invalid';
    END IF;
    WITH repaired AS (
        SELECT subject.id,
               iam_profile_installation_legacy_document_v1(
                   subject.profile_values, subject.profile_sha256,
                   snapshot.field_id,
                   snapshot.field_key, snapshot.value_kind,
                   snapshot.cardinality, snapshot.field_state
               ) AS canonical_json
        FROM iam_users AS subject
        JOIN iam_profile_field_revisions AS snapshot
          ON snapshot.schema_version_id = subject.profile_schema_version_id
         AND snapshot.resource_kind = 'user'
         AND snapshot.field_key = 'display_name'
        WHERE subject.profile_values ? 'display_name'
    )
    UPDATE iam_users AS subject
    SET profile_values = repaired.canonical_json::jsonb,
        profile_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(repaired.canonical_json, 'UTF8')
        ),
        revision = subject.revision + 1,
        updated_at = pg_catalog.statement_timestamp()
    FROM repaired
    WHERE subject.id = repaired.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    v_changed := v_changed + v_count;
    WITH repaired AS (
        SELECT subject.id,
               iam_profile_installation_legacy_document_v1(
                   subject.profile_values, subject.profile_sha256,
                   snapshot.field_id,
                   snapshot.field_key, snapshot.value_kind,
                   snapshot.cardinality, snapshot.field_state
               ) AS canonical_json
        FROM iam_organizations AS subject
        JOIN iam_profile_field_revisions AS snapshot
          ON snapshot.schema_version_id = subject.profile_schema_version_id
         AND snapshot.resource_kind = 'organization'
         AND snapshot.field_key = 'display_name'
        WHERE subject.profile_values ? 'display_name'
    )
    UPDATE iam_organizations AS subject
    SET profile_values = repaired.canonical_json::jsonb,
        profile_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(repaired.canonical_json, 'UTF8')
        ),
        revision = subject.revision + 1,
        updated_at = pg_catalog.statement_timestamp()
    FROM repaired
    WHERE subject.id = repaired.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    v_changed := v_changed + v_count;
    WITH repaired AS (
        SELECT subject.id,
               iam_profile_installation_legacy_document_v1(
                   subject.profile_values, subject.profile_sha256,
                   snapshot.field_id,
                   snapshot.field_key, snapshot.value_kind,
                   snapshot.cardinality, snapshot.field_state
               ) AS canonical_json
        FROM iam_roles AS subject
        JOIN iam_profile_field_revisions AS snapshot
          ON snapshot.schema_version_id = subject.profile_schema_version_id
         AND snapshot.resource_kind = 'role'
         AND snapshot.field_key = 'display_name'
        WHERE subject.profile_values ? 'display_name'
    )
    UPDATE iam_roles AS subject
    SET profile_values = repaired.canonical_json::jsonb,
        profile_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(repaired.canonical_json, 'UTF8')
        ),
        revision = subject.revision + 1,
        updated_at = pg_catalog.statement_timestamp()
    FROM repaired
    WHERE subject.id = repaired.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_changed + v_count;
END;
$function$;
SELECT iam_profile_display_name_key_repair_v1('display_name');
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_profile_installation_legacy_document_v1(
    jsonb,bytea,uuid,text,text,text,text
)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_profile_display_name_key_repair_v1(text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_profile_display_name_key_repair_v1(text)
    FROM forgecrud_runtime;
