-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_preconversion_reserve_v1(
    p_scope_schema name, p_plan_id uuid, p_resource_kind text,
    p_resource_id uuid, p_actual_source_schema_id uuid,
    p_source_revision bigint, p_source_sha256 bytea,
    p_planned_target_sha256 bytea, p_first_target_sha256 bytea
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_page_number integer;
    v_source_version bigint;
    v_current_schema_id uuid;
    v_current_revision bigint;
    v_current_sha256 bytea;
    v_existing_page_number integer;
    v_existing_resource_kind text;
    v_existing_schema_id uuid;
    v_existing_revision bigint;
    v_existing_source_sha256 bytea;
    v_existing_target_sha256 bytea;
    v_existing_first_sha256 bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_plan_id IS NULL OR p_resource_kind NOT IN ('user','organization','role') OR
       p_resource_id IS NULL OR p_actual_source_schema_id IS NULL OR
       p_source_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_planned_target_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_first_target_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile preconversion reservation is invalid';
    END IF;
    SELECT page.page_number, source.version
    INTO v_page_number, v_source_version
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_backfill_pages AS page ON page.plan_id = plan.id
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind AND target.state = 'draft'
    WHERE plan.id = p_plan_id AND plan.resource_kind = p_resource_kind
      AND plan.state IN ('planned','running') AND page.state = 'queued'
      AND (page.after_resource_id IS NULL OR p_resource_id > page.after_resource_id)
      AND p_resource_id <= page.through_resource_id
    FOR SHARE OF plan, page, source, target;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile preconversion page is stale';
    END IF;
    IF p_resource_kind = 'user' THEN
        SELECT resource.profile_schema_version_id, resource.revision,
               resource.profile_sha256
        INTO v_current_schema_id, v_current_revision, v_current_sha256
        FROM iam_users AS resource WHERE resource.id = p_resource_id
        FOR UPDATE;
    ELSIF p_resource_kind = 'organization' THEN
        SELECT resource.profile_schema_version_id, resource.revision,
               resource.profile_sha256
        INTO v_current_schema_id, v_current_revision, v_current_sha256
        FROM iam_organizations AS resource WHERE resource.id = p_resource_id
        FOR UPDATE;
    ELSE
        SELECT resource.profile_schema_version_id, resource.revision,
               resource.profile_sha256
        INTO v_current_schema_id, v_current_revision, v_current_sha256
        FROM iam_roles AS resource WHERE resource.id = p_resource_id
        FOR UPDATE;
    END IF;
    IF v_current_schema_id IS DISTINCT FROM p_actual_source_schema_id OR
       v_current_revision IS DISTINCT FROM p_source_revision OR
       v_current_sha256 IS DISTINCT FROM p_source_sha256 OR NOT EXISTS (
           SELECT 1 FROM iam_profile_schema_versions AS actual
           WHERE actual.id = p_actual_source_schema_id
             AND actual.resource_kind = p_resource_kind
             AND actual.state IN ('active','retired')
             AND actual.version <= v_source_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile preconversion source changed';
    END IF;
    SELECT receipt.page_number, receipt.resource_kind,
           receipt.actual_source_schema_id, receipt.source_revision,
           receipt.source_sha256, receipt.planned_target_sha256,
           receipt.first_target_sha256
    INTO v_existing_page_number, v_existing_resource_kind,
         v_existing_schema_id, v_existing_revision,
         v_existing_source_sha256, v_existing_target_sha256,
         v_existing_first_sha256
    FROM iam_profile_backfill_preconversions AS receipt
    WHERE receipt.plan_id = p_plan_id AND receipt.resource_id = p_resource_id;
    IF FOUND THEN
        IF v_existing_page_number IS DISTINCT FROM v_page_number OR
           v_existing_resource_kind IS DISTINCT FROM p_resource_kind OR
           v_existing_schema_id IS DISTINCT FROM p_actual_source_schema_id OR
           v_existing_revision IS DISTINCT FROM p_source_revision OR
           v_existing_source_sha256 IS DISTINCT FROM p_source_sha256 OR
           v_existing_target_sha256 IS DISTINCT FROM p_planned_target_sha256 OR
           v_existing_first_sha256 IS DISTINCT FROM p_first_target_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM profile preconversion identity was reused';
        END IF;
        RETURN v_existing_page_number;
    END IF;
    INSERT INTO iam_profile_backfill_preconversions (
        plan_id, page_number, resource_id, resource_kind,
        actual_source_schema_id, source_revision, source_sha256,
        planned_target_sha256, first_target_revision,
        first_target_sha256, created_at
    ) VALUES (
        p_plan_id, v_page_number, p_resource_id, p_resource_kind,
        p_actual_source_schema_id, p_source_revision, p_source_sha256,
        p_planned_target_sha256, p_source_revision + 1,
        p_first_target_sha256, pg_catalog.transaction_timestamp()
    );
    RETURN v_page_number;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_preconversion_reserve_v1(
    name,uuid,text,uuid,uuid,bigint,bytea,bytea,bytea
) FROM PUBLIC;
