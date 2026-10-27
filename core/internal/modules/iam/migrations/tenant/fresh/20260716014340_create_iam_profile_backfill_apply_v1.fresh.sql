-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_backfill_apply_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer, p_resource_id uuid,
    p_expected_revision bigint, p_source_sha256 bytea,
    p_target_document bytea, p_target_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_plan_id uuid;
    v_resource_kind text;
    v_source_schema_id uuid;
    v_target_schema_id uuid;
    v_page_job_id uuid;
    v_after_resource_id uuid;
    v_through_resource_id uuid;
    v_values jsonb;
    v_revision bigint;
    v_current_schema uuid;
    v_current_sha bytea;
BEGIN
    IF NOT iam_maintenance_claim_assert_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.profile.backfill', p_payload_revision,
        p_payload_reference, p_payload_sha256
    ) OR p_plan_id IS NULL OR p_page_number < 1 OR p_resource_id IS NULL OR
       p_expected_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_target_document) NOT BETWEEN 2 AND 1048576 OR
       pg_catalog.octet_length(p_target_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(p_target_document) IS DISTINCT FROM p_target_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill apply input is invalid';
    END IF;
    SELECT plan.id, plan.resource_kind, plan.source_schema_id,
           plan.target_schema_id
    INTO v_plan_id, v_resource_kind, v_source_schema_id, v_target_schema_id
    FROM iam_profile_backfill_plans AS plan
    WHERE plan.id = p_plan_id AND plan.state IN ('planned', 'running')
    FOR UPDATE;
    SELECT page.job_id, page.after_resource_id, page.through_resource_id
    INTO v_page_job_id, v_after_resource_id, v_through_resource_id
    FROM iam_profile_backfill_pages AS page
    WHERE page.plan_id = p_plan_id AND page.page_number = p_page_number
      AND page.job_id = p_job_id AND page.state = 'queued'
    FOR UPDATE;
    IF v_plan_id IS NULL OR v_page_job_id IS NULL OR
       (v_after_resource_id IS NOT NULL AND
        p_resource_id <= v_after_resource_id) OR
       p_resource_id > v_through_resource_id THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile backfill plan or page is stale';
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
          )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS required
        WHERE required.schema_version_id = v_target_schema_id
          AND required.resource_kind = v_resource_kind
          AND required.field_state = 'active' AND required.required
          AND NOT v_values ? required.field_id::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile backfill target document is invalid';
    END IF;
    IF v_resource_kind = 'user' THEN
        SELECT profile_schema_version_id, profile_sha256
        INTO v_current_schema, v_current_sha
        FROM iam_users WHERE id = p_resource_id AND revision = p_expected_revision
        FOR UPDATE;
    ELSIF v_resource_kind = 'organization' THEN
        SELECT profile_schema_version_id, profile_sha256
        INTO v_current_schema, v_current_sha
        FROM iam_organizations WHERE id = p_resource_id AND revision = p_expected_revision
        FOR UPDATE;
    ELSE
        SELECT profile_schema_version_id, profile_sha256
        INTO v_current_schema, v_current_sha
        FROM iam_roles WHERE id = p_resource_id AND revision = p_expected_revision
        FOR UPDATE;
    END IF;
    IF v_current_schema IS DISTINCT FROM v_source_schema_id OR
       v_current_sha IS DISTINCT FROM p_source_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile backfill source changed';
    END IF;
    IF v_resource_kind = 'user' THEN
        UPDATE iam_users SET profile_schema_version_id = v_target_schema_id,
            profile_values = v_values, profile_sha256 = p_target_sha256,
            revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
        WHERE id = p_resource_id RETURNING revision INTO v_revision;
    ELSIF v_resource_kind = 'organization' THEN
        UPDATE iam_organizations SET profile_schema_version_id = v_target_schema_id,
            profile_values = v_values, profile_sha256 = p_target_sha256,
            revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
        WHERE id = p_resource_id RETURNING revision INTO v_revision;
    ELSE
        UPDATE iam_roles SET profile_schema_version_id = v_target_schema_id,
            profile_values = v_values, profile_sha256 = p_target_sha256,
            revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
        WHERE id = p_resource_id RETURNING revision INTO v_revision;
    END IF;
    UPDATE iam_profile_backfill_plans SET state = 'running',
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_plan_id AND state = 'planned';
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_apply_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,bigint,bytea,bytea,bytea
) FROM PUBLIC;
