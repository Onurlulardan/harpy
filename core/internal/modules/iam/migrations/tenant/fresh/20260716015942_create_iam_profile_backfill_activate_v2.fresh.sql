-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_backfill_activate_v2(
    p_plan_id uuid, p_resource_kind text, p_source_schema_id uuid,
    p_target_schema_id uuid, p_source_count bigint, p_now timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_total_count bigint;
    v_target_count bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_plan_id IS NULL OR p_resource_kind NOT IN ('user','organization','role') OR
       p_source_schema_id IS NULL OR p_target_schema_id IS NULL OR
       p_source_schema_id = p_target_schema_id OR p_source_count < 0 OR
       p_now IS NULL OR NOT EXISTS (
           SELECT 1 FROM iam_profile_backfill_plans AS plan
           JOIN iam_profile_schema_versions AS source
             ON source.id = plan.source_schema_id
            AND source.resource_kind = plan.resource_kind
            AND source.state = 'active'
           JOIN iam_profile_schema_versions AS target
             ON target.id = plan.target_schema_id
            AND target.resource_kind = plan.resource_kind
            AND target.state = 'draft'
            AND target.version = source.version + 1
           WHERE plan.id = p_plan_id AND plan.resource_kind = p_resource_kind
             AND plan.source_schema_id = p_source_schema_id
             AND plan.target_schema_id = p_target_schema_id
             AND plan.source_count = p_source_count
             AND plan.state IN ('planned','running')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile activation compare-and-swap failed';
    END IF;
    SELECT pg_catalog.count(*),
           pg_catalog.count(*) FILTER (
               WHERE resource.profile_schema_version_id = p_target_schema_id
           )
    INTO v_total_count, v_target_count
    FROM (
        SELECT usr.profile_schema_version_id FROM iam_users AS usr
        WHERE p_resource_kind = 'user'
        UNION ALL
        SELECT organization.profile_schema_version_id
        FROM iam_organizations AS organization
        WHERE p_resource_kind = 'organization'
        UNION ALL
        SELECT role.profile_schema_version_id FROM iam_roles AS role
        WHERE p_resource_kind = 'role'
    ) AS resource;
    IF v_total_count <> v_target_count OR v_target_count < p_source_count THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile activation row lineage is inconsistent';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT * FROM iam_profile_fields
            WHERE resource_kind = p_resource_kind
        ) AS field
        FULL JOIN (
            SELECT * FROM iam_profile_field_revisions
            WHERE schema_version_id = p_target_schema_id
              AND resource_kind = p_resource_kind
        ) AS snapshot ON snapshot.field_id = field.id
        WHERE field.id IS NULL OR snapshot.field_id IS NULL OR
              field.field_key <> snapshot.field_key OR
              field.value_kind <> snapshot.value_kind OR
              field.cardinality <> snapshot.cardinality OR
              field.required <> snapshot.required OR
              field.classification <> snapshot.classification OR
              field.normalization_profile <> snapshot.normalization_profile OR
              (field.state = 'retired' AND snapshot.field_state = 'active')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile activation field lineage is inconsistent';
    END IF;
    UPDATE iam_profile_fields AS field
    SET state = 'retired', retired_at = p_now, updated_at = p_now,
        revision = field.revision + 1
    FROM iam_profile_field_revisions AS snapshot
    WHERE snapshot.schema_version_id = p_target_schema_id
      AND snapshot.resource_kind = p_resource_kind
      AND snapshot.field_id = field.id
      AND field.resource_kind = p_resource_kind
      AND snapshot.field_state = 'retired' AND field.state = 'active';
    UPDATE iam_profile_schema_versions
    SET state = 'retired', retired_at = p_now, updated_at = p_now,
        revision = revision + 1
    WHERE id = p_source_schema_id AND resource_kind = p_resource_kind
      AND state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile source pointer changed';
    END IF;
    UPDATE iam_profile_schema_versions
    SET state = 'active', activated_at = p_now, updated_at = p_now,
        revision = revision + 1
    WHERE id = p_target_schema_id AND resource_kind = p_resource_kind
      AND state = 'draft';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile target pointer changed';
    END IF;
    UPDATE iam_profile_backfill_plans
    SET state = 'activated', activated_at = p_now,
        revision = revision + 1, updated_at = p_now
    WHERE id = p_plan_id AND state IN ('planned','running');
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile plan activation changed';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_backfill_activate_v2(
    uuid,text,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
