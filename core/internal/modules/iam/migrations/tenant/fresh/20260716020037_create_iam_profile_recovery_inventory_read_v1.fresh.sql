-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_inventory_open_v1(
    p_scope_schema name, p_parent_plan_id uuid,
    p_actor_id uuid, p_manage_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS refcursor
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_kind text;
    v_source_schema_id uuid;
    v_target_schema_id uuid;
    v_source_version bigint;
    v_source_definition_sha256 bytea;
    v_target_definition_sha256 bytea;
    v_cursor refcursor;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_parent_plan_id IS NULL OR
       NOT iam_profile_actor_assert_v1(
           p_scope_schema, p_actor_id, p_manage_permission_id,
           p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile recovery inventory authority is denied';
    END IF;
    SELECT parent.resource_kind, parent.source_schema_id, parent.target_schema_id,
           source.version, source.definition_sha256, target.definition_sha256
    INTO v_resource_kind, v_source_schema_id, v_target_schema_id,
         v_source_version, v_source_definition_sha256,
         v_target_definition_sha256
    FROM iam_profile_backfill_plans AS parent
    JOIN iam_profile_schema_versions AS source
      ON source.id = parent.source_schema_id
     AND source.resource_kind = parent.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = parent.target_schema_id
     AND target.resource_kind = parent.resource_kind AND target.state = 'draft'
     AND target.version = source.version + 1
    WHERE parent.id = p_parent_plan_id AND parent.state = 'fenced'
      AND NOT EXISTS (
          SELECT 1 FROM iam_profile_backfill_plans AS child
          WHERE child.recovery_of_plan_id = parent.id
      )
    FOR SHARE OF parent, source, target;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile recovery parent is unavailable';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM iam_profile_recovery_resources_v1 AS resource
        LEFT JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.schema_version_id
         AND row_schema.resource_kind = resource.resource_kind
        WHERE resource.resource_kind = v_resource_kind AND (
            row_schema.id IS NULL OR
            (resource.schema_version_id = v_target_schema_id AND
             row_schema.state <> 'draft') OR
            (resource.schema_version_id <> v_target_schema_id AND
             (row_schema.state NOT IN ('active', 'retired') OR
              row_schema.version > v_source_version))
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery inventory lineage is invalid';
    END IF;
    IF iam_profile_schema_snapshot_json_v1(v_source_schema_id) IS NULL OR
       iam_profile_schema_snapshot_json_v1(v_target_schema_id) IS NULL OR
       iam_profile_definition_sha256_v1(v_source_schema_id) IS DISTINCT FROM
           v_source_definition_sha256 OR
       iam_profile_definition_sha256_v1(v_target_schema_id) IS DISTINCT FROM
           v_target_definition_sha256 OR EXISTS (
           SELECT 1 FROM (
               SELECT DISTINCT resource.schema_version_id
               FROM iam_profile_recovery_resources_v1 AS resource
               WHERE resource.resource_kind = v_resource_kind
           ) AS inventory_schema
           JOIN iam_profile_schema_versions AS row_schema
             ON row_schema.id = inventory_schema.schema_version_id
            AND row_schema.resource_kind = v_resource_kind
           WHERE iam_profile_definition_sha256_v1(row_schema.id) IS DISTINCT FROM
                   row_schema.definition_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery schema snapshot drifted';
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_profile_backfill_mappings AS mapping
        LEFT JOIN iam_profile_field_revisions AS source_field
          ON source_field.schema_version_id = v_source_schema_id
         AND source_field.resource_kind = v_resource_kind
         AND source_field.field_id = mapping.source_field_id
        LEFT JOIN iam_profile_field_revisions AS target_field
          ON target_field.schema_version_id = v_target_schema_id
         AND target_field.resource_kind = v_resource_kind
         AND target_field.field_id = mapping.target_field_id
        WHERE mapping.plan_id = p_parent_plan_id AND (
            source_field.field_id IS NULL OR target_field.field_id IS NULL OR
            mapping.semantic_sha256 <> target_field.semantic_sha256 OR
            (mapping.converter_kind = 'copy_v1' AND
             source_field.semantic_sha256 <> target_field.semantic_sha256)
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery mapping semantic drifted';
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_profile_recovery_resources_v1 AS resource
        WHERE resource.resource_kind = v_resource_kind
          AND resource.schema_version_id = v_target_schema_id AND (
              EXISTS (
                  SELECT 1 FROM pg_catalog.jsonb_each(resource.profile_values)
                       AS entry(key, value)
                  WHERE entry.key !~
                      '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
                     OR pg_catalog.jsonb_typeof(entry.value) = 'null'
                     OR NOT EXISTS (
                         SELECT 1 FROM iam_profile_field_revisions AS field
                         WHERE field.schema_version_id = v_target_schema_id
                           AND field.resource_kind = v_resource_kind
                           AND field.field_id::text = entry.key
                     )
              ) OR EXISTS (
                  SELECT 1 FROM iam_profile_field_revisions AS required
                  WHERE required.schema_version_id = v_target_schema_id
                    AND required.resource_kind = v_resource_kind
                    AND required.field_state = 'active' AND required.required
                    AND NOT resource.profile_values ? required.field_id::text
              )
          )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery target document is invalid';
    END IF;
    v_cursor := 'iam_profile_recovery_inventory_v1';
    IF v_resource_kind = 'user' THEN
        OPEN v_cursor NO SCROLL FOR
        SELECT resource.id,resource.profile_schema_version_id,row_schema.version,
               row_schema.state,CASE WHEN resource.profile_schema_version_id =
               v_target_schema_id THEN 'target' ELSE 'source' END,
               resource.revision,resource.profile_values::text,resource.profile_sha256
        FROM iam_users AS resource JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.profile_schema_version_id
         AND row_schema.resource_kind = 'user' ORDER BY resource.id;
    ELSIF v_resource_kind = 'organization' THEN
        OPEN v_cursor NO SCROLL FOR
        SELECT resource.id,resource.profile_schema_version_id,row_schema.version,
               row_schema.state,CASE WHEN resource.profile_schema_version_id =
               v_target_schema_id THEN 'target' ELSE 'source' END,
               resource.revision,resource.profile_values::text,resource.profile_sha256
        FROM iam_organizations AS resource JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.profile_schema_version_id
         AND row_schema.resource_kind = 'organization' ORDER BY resource.id;
    ELSE
        OPEN v_cursor NO SCROLL FOR
        SELECT resource.id,resource.profile_schema_version_id,row_schema.version,
               row_schema.state,CASE WHEN resource.profile_schema_version_id =
               v_target_schema_id THEN 'target' ELSE 'source' END,
               resource.revision,resource.profile_values::text,resource.profile_sha256
        FROM iam_roles AS resource JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.profile_schema_version_id
         AND row_schema.resource_kind = 'role' ORDER BY resource.id;
    END IF;
    RETURN v_cursor;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_inventory_open_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
