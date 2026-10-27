-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_schema_read_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_schema_id uuid
)
RETURNS TABLE (
    schema_id uuid, resource_kind text, schema_version bigint,
    schema_state text, definition_sha256 bytea, field_id uuid,
    field_key text, value_kind text, cardinality text, required boolean,
    classification text, normalization_profile text, field_state text,
    validation_schema text, validation_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_kind text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.profile.backfill', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile backfill schema authority is invalid';
    END IF;
    SELECT plan.resource_kind INTO v_resource_kind
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_backfill_pages AS page ON page.plan_id = plan.id
    WHERE plan.id = p_plan_id AND page.job_id = p_job_id
      AND page.payload_reference = p_payload_reference
      AND page.payload_sha256 = p_payload_sha256 AND page.state = 'queued'
      AND p_schema_id IN (plan.source_schema_id, plan.target_schema_id);
    IF v_resource_kind IS NULL OR (
        SELECT pg_catalog.count(*) FROM iam_profile_field_revisions AS snapshot
        WHERE snapshot.schema_version_id = p_schema_id
          AND snapshot.resource_kind = v_resource_kind
    ) > 250 THEN
        RAISE EXCEPTION USING ERRCODE = '54000',
            MESSAGE = 'IAM profile backfill schema bound is invalid';
    END IF;
    RETURN QUERY
    SELECT version.id, version.resource_kind, version.version, version.state,
           version.definition_sha256, field.id, field.field_key,
           field.value_kind, field.cardinality, field.required,
           field.classification, field.normalization_profile, field.state,
           snapshot.validation_schema::text, snapshot.validation_sha256
    FROM iam_profile_schema_versions AS version
    LEFT JOIN iam_profile_field_revisions AS snapshot
      ON snapshot.schema_version_id = version.id
     AND snapshot.resource_kind = version.resource_kind
    LEFT JOIN iam_profile_fields AS field
      ON field.id = snapshot.field_id
     AND field.resource_kind = snapshot.resource_kind
    WHERE version.id = p_schema_id AND version.resource_kind = v_resource_kind
    ORDER BY field.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_schema_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,uuid
) FROM PUBLIC;
