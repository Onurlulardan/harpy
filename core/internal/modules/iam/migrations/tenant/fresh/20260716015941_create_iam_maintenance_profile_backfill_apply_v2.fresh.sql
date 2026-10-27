-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_backfill_apply_v2(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer, p_resource_id uuid,
    p_actual_schema_version_id uuid, p_expected_revision bigint,
    p_source_sha256 bytea, p_target_document bytea, p_target_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_kind text;
    v_source_schema_id uuid;
    v_source_version bigint;
    v_target_schema_id uuid;
    v_after_resource_id uuid;
    v_through_resource_id uuid;
    v_source_values jsonb;
    v_values jsonb;
    v_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.profile.backfill', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) OR p_plan_id IS NULL OR p_page_number < 1 OR p_resource_id IS NULL OR
       p_actual_schema_version_id IS NULL OR p_expected_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_target_document) NOT BETWEEN 2 AND 1048576 OR
       pg_catalog.octet_length(p_target_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(p_target_document) IS DISTINCT FROM p_target_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile lineage apply input is invalid';
    END IF;
    SELECT plan.resource_kind, plan.source_schema_id, source.version,
           plan.target_schema_id, page.after_resource_id,
           page.through_resource_id
    INTO v_resource_kind, v_source_schema_id, v_source_version,
         v_target_schema_id, v_after_resource_id, v_through_resource_id
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_backfill_pages AS page ON page.plan_id = plan.id
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind AND target.state = 'draft'
    WHERE plan.id = p_plan_id AND plan.state IN ('planned', 'running')
      AND page.page_number = p_page_number AND page.job_id = p_job_id
      AND page.state = 'queued'
    FOR UPDATE OF plan, page, source, target;
    IF v_resource_kind IS NULL OR
       (v_after_resource_id IS NOT NULL AND p_resource_id <= v_after_resource_id) OR
       p_resource_id > v_through_resource_id OR NOT EXISTS (
           SELECT 1 FROM iam_profile_schema_versions AS actual
           WHERE actual.id = p_actual_schema_version_id
             AND actual.resource_kind = v_resource_kind
             AND actual.state IN ('active', 'retired')
             AND actual.version <= v_source_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile lineage apply is stale';
    END IF;
    IF v_resource_kind = 'user' THEN
        SELECT usr.profile_values INTO v_source_values
        FROM iam_users AS usr WHERE usr.id = p_resource_id
          AND usr.profile_schema_version_id = p_actual_schema_version_id
          AND usr.revision = p_expected_revision
          AND usr.profile_sha256 = p_source_sha256 FOR UPDATE;
    ELSIF v_resource_kind = 'organization' THEN
        SELECT organization.profile_values INTO v_source_values
        FROM iam_organizations AS organization
        WHERE organization.id = p_resource_id
          AND organization.profile_schema_version_id = p_actual_schema_version_id
          AND organization.revision = p_expected_revision
          AND organization.profile_sha256 = p_source_sha256 FOR UPDATE;
    ELSE
        SELECT role.profile_values INTO v_source_values
        FROM iam_roles AS role WHERE role.id = p_resource_id
          AND role.profile_schema_version_id = p_actual_schema_version_id
          AND role.revision = p_expected_revision
          AND role.profile_sha256 = p_source_sha256 FOR UPDATE;
    END IF;
    IF v_source_values IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile lineage source changed';
    END IF;
    v_values := pg_catalog.convert_from(p_target_document, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_values) <> 'object' OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_each(v_values) AS entry(key, value)
        WHERE entry.key !~
            '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
          OR pg_catalog.jsonb_typeof(entry.value) = 'null'
          OR NOT EXISTS (
              SELECT 1 FROM iam_profile_field_revisions AS field
              WHERE field.schema_version_id = v_target_schema_id
                AND field.resource_kind = v_resource_kind
                AND field.field_id = entry.key::uuid
                AND (field.field_state = 'active' OR (
                    field.field_state = 'retired'
                    AND v_source_values ? entry.key
                    AND v_source_values -> entry.key = entry.value
                    AND EXISTS (
                        SELECT 1 FROM iam_profile_field_revisions AS source_field
                        WHERE source_field.schema_version_id =
                              p_actual_schema_version_id
                          AND source_field.resource_kind = v_resource_kind
                          AND source_field.field_id = field.field_id
                          AND source_field.semantic_sha256 = field.semantic_sha256
                    )
                ))
          )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS required
        WHERE required.schema_version_id = v_target_schema_id
          AND required.resource_kind = v_resource_kind
          AND required.field_state = 'active' AND required.required
          AND NOT v_values ? required.field_id::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile lineage target document is invalid';
    END IF;
    IF v_resource_kind = 'user' THEN
        UPDATE iam_users SET profile_schema_version_id = v_target_schema_id,
            profile_values = v_values, profile_sha256 = p_target_sha256,
            revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
        WHERE id = p_resource_id
          AND profile_schema_version_id = p_actual_schema_version_id
          AND revision = p_expected_revision AND profile_sha256 = p_source_sha256
        RETURNING revision INTO v_revision;
    ELSIF v_resource_kind = 'organization' THEN
        UPDATE iam_organizations SET profile_schema_version_id = v_target_schema_id,
            profile_values = v_values, profile_sha256 = p_target_sha256,
            revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
        WHERE id = p_resource_id
          AND profile_schema_version_id = p_actual_schema_version_id
          AND revision = p_expected_revision AND profile_sha256 = p_source_sha256
        RETURNING revision INTO v_revision;
    ELSE
        UPDATE iam_roles SET profile_schema_version_id = v_target_schema_id,
            profile_values = v_values, profile_sha256 = p_target_sha256,
            revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
        WHERE id = p_resource_id
          AND profile_schema_version_id = p_actual_schema_version_id
          AND revision = p_expected_revision AND profile_sha256 = p_source_sha256
        RETURNING revision INTO v_revision;
    END IF;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile lineage source changed';
    END IF;
    UPDATE iam_profile_backfill_plans SET state = 'running',
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_plan_id AND state = 'planned';
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_apply_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) FROM PUBLIC;
