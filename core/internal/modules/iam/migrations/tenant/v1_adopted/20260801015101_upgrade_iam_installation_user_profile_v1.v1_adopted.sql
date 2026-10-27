-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_user_profile_upgrade_v1(p_operation_id uuid)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_source_id uuid; v_target_id uuid; v_display_id uuid;
    v_first_id uuid; v_last_id uuid; v_phone_id uuid;
    v_source_version bigint;
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_validation jsonb;
BEGIN
    v_source_id := iam_installation_derived_id_v1(p_operation_id, 'user-profile-schema');
    v_target_id := iam_installation_derived_id_v1(p_operation_id, 'user-profile-schema-v2');
    v_display_id := iam_installation_derived_id_v1(p_operation_id, 'user-display-name-field');
    v_first_id := iam_installation_derived_id_v1(p_operation_id, 'user-first-name-field');
    v_last_id := iam_installation_derived_id_v1(p_operation_id, 'user-last-name-field');
    v_phone_id := iam_installation_derived_id_v1(p_operation_id, 'user-phone-field');
    IF EXISTS (
        SELECT 1 FROM iam_profile_schema_versions AS target
        WHERE target.id = v_target_id AND target.resource_kind = 'user'
          AND target.state = 'active'
    ) THEN
        IF (SELECT pg_catalog.count(*)
            FROM iam_profile_field_revisions AS field
            WHERE field.schema_version_id = v_target_id
              AND field.resource_kind = 'user'
              AND field.field_state = 'active'
              AND field.field_id IN (
                  v_display_id, v_first_id, v_last_id, v_phone_id
              )) IS DISTINCT FROM 4 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM installation user profile successor is invalid';
        END IF;
        RETURN true;
    END IF;
    SELECT source.version INTO v_source_version
    FROM iam_profile_schema_versions AS source
    WHERE source.id = v_source_id AND source.resource_kind = 'user'
      AND source.state = 'active'
    FOR UPDATE;
    IF NOT FOUND THEN
        IF EXISTS (
            SELECT 1
            FROM iam_profile_schema_versions AS active
            JOIN iam_profile_field_revisions AS field
              ON field.schema_version_id = active.id
             AND field.resource_kind = active.resource_kind
            WHERE active.resource_kind = 'user' AND active.state = 'active'
              AND field.field_state = 'active'
              AND field.field_key IN ('first_name','last_name','phone')
            GROUP BY active.id HAVING pg_catalog.count(*) = 3
        ) THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM installation user profile source is unavailable';
    END IF;
    IF (SELECT pg_catalog.count(*)
        FROM iam_profile_field_revisions AS field
        WHERE field.schema_version_id = v_source_id
          AND field.resource_kind = 'user') IS DISTINCT FROM 1
       OR NOT EXISTS (
           SELECT 1 FROM iam_profile_field_revisions AS field
           WHERE field.schema_version_id = v_source_id
             AND field.resource_kind = 'user'
             AND field.field_id = v_display_id
             AND field.field_key = 'display_name'
             AND field.value_kind = 'string'
             AND field.cardinality = 'single'
             AND field.required AND field.classification = 'public'
             AND field.normalization_profile = 'none_v1'
             AND field.field_state = 'active'
       )
       OR EXISTS (
           SELECT 1 FROM iam_users AS subject
           WHERE subject.profile_schema_version_id = v_source_id
             AND (
                 pg_catalog.jsonb_typeof(subject.profile_values) <> 'object'
                 OR (SELECT pg_catalog.count(*)
                     FROM pg_catalog.jsonb_object_keys(subject.profile_values)) <> 1
                 OR NOT subject.profile_values ? v_display_id::text
                 OR pg_catalog.jsonb_typeof(
                     subject.profile_values -> v_display_id::text
                 ) <> 'string'
             )
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM installation user profile source is invalid';
    END IF;
    v_validation := '{"maximum_bytes":200}'::jsonb;
    INSERT INTO iam_profile_fields (
        id, resource_kind, field_key, value_kind, cardinality, required,
        classification, normalization_profile, state, created_at,
        updated_at, retired_at, revision
    ) VALUES
        (v_first_id,'user','first_name','string','single',false,
         'public','none_v1','active',v_now,v_now,NULL,1),
        (v_last_id,'user','last_name','string','single',false,
         'public','none_v1','active',v_now,v_now,NULL,1),
        (v_phone_id,'user','phone','string','single',false,
         'public','none_v1','active',v_now,v_now,NULL,1);
    INSERT INTO iam_profile_schema_versions (
        id, resource_kind, version, state, definition_sha256,
        activated_at, retired_at, created_at, updated_at, revision
    ) VALUES (
        v_target_id, 'user', v_source_version + 1, 'draft',
        pg_catalog.sha256(pg_catalog.uuid_send(v_target_id)),
        NULL, NULL, v_now, v_now, 1
    );
    INSERT INTO iam_profile_field_revisions
    SELECT v_target_id, resource_kind, field_id, validation_schema,
        validation_sha256, v_now, field_key, value_kind, cardinality,
        required, classification, normalization_profile, field_state,
        semantic_sha256
    FROM iam_profile_field_revisions
    WHERE schema_version_id = v_source_id AND field_id = v_display_id;
    INSERT INTO iam_profile_field_revisions (
        schema_version_id, resource_kind, field_id, validation_schema,
        validation_sha256, created_at, field_key, value_kind, cardinality,
        required, classification, normalization_profile, field_state,
        semantic_sha256
    )
    SELECT v_target_id, 'user', field.id,
        CASE WHEN field.id = v_phone_id
             THEN '{"maximum_bytes":32}'::jsonb ELSE v_validation END,
        pg_catalog.sha256(pg_catalog.convert_to(
            iam_profile_validation_canonical_v1(
                CASE WHEN field.id = v_phone_id
                     THEN '{"maximum_bytes":32}'::jsonb ELSE v_validation END
            ), 'UTF8'
        )), v_now, field.field_key, field.value_kind, field.cardinality,
        field.required, field.classification, field.normalization_profile,
        field.state, iam_profile_semantic_sha256_v1(
            field.id, field.resource_kind, field.field_key, field.value_kind,
            field.cardinality, field.required, field.classification,
            field.normalization_profile
        )
    FROM iam_profile_fields AS field
    WHERE field.id IN (v_first_id, v_last_id, v_phone_id);
    UPDATE iam_profile_schema_versions
    SET definition_sha256 = iam_profile_definition_sha256_v1(v_target_id),
        updated_at = v_now, revision = 2
    WHERE id = v_target_id AND state = 'draft' AND revision = 1;
    WITH documents AS (
        SELECT subject.id,
               '{' || pg_catalog.string_agg(
                   pg_catalog.to_json(entry.key)::text || ':' ||
                   entry.value::text, ',' ORDER BY entry.key
               ) || '}' AS canonical_json
        FROM iam_users AS subject
        CROSS JOIN LATERAL pg_catalog.jsonb_each(
            subject.profile_values ||
            pg_catalog.jsonb_build_object(
                v_first_id::text,
                subject.profile_values -> v_display_id::text
            )
        ) AS entry(key,value)
        WHERE subject.profile_schema_version_id = v_source_id
        GROUP BY subject.id
    )
    UPDATE iam_users AS subject
    SET profile_schema_version_id = v_target_id,
        profile_values = documents.canonical_json::jsonb,
        profile_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(documents.canonical_json, 'UTF8')
        ),
        revision = subject.revision + 1, updated_at = v_now
    FROM documents WHERE subject.id = documents.id;
    UPDATE iam_profile_schema_versions
    SET state = 'retired', retired_at = v_now, updated_at = v_now,
        revision = revision + 1
    WHERE id = v_source_id AND state = 'active';
    UPDATE iam_profile_schema_versions
    SET state = 'active', activated_at = v_now, updated_at = v_now,
        revision = revision + 1
    WHERE id = v_target_id AND state = 'draft' AND revision = 2;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM installation user profile activation conflict';
    END IF;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_user_profile_upgrade_v1(uuid) FROM PUBLIC;
