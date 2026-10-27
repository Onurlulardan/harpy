-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_lineage_resources_read_v2(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer
)
RETURNS TABLE (
    resource_id uuid, actual_schema_version_id uuid,
    source_revision bigint, profile_values text, profile_sha256 bytea,
    preconverted boolean, planned_target_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_kind text;
    v_source_version bigint;
    v_after_resource_id uuid;
    v_through_resource_id uuid;
    v_expected_count integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.profile.backfill', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile lineage read authority is invalid';
    END IF;
    SELECT plan.resource_kind, source.version,
           page.after_resource_id, page.through_resource_id,
           page.expected_count
    INTO v_resource_kind, v_source_version, v_after_resource_id,
         v_through_resource_id, v_expected_count
    FROM iam_profile_backfill_pages AS page
    JOIN iam_profile_backfill_plans AS plan ON plan.id = page.plan_id
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind
     AND source.state = 'active'
    WHERE page.plan_id = p_plan_id AND page.page_number = p_page_number
      AND page.job_id = p_job_id
      AND page.payload_reference = p_payload_reference
      AND page.payload_sha256 = p_payload_sha256
      AND page.state = 'queued' AND plan.state IN ('planned', 'running');
    IF v_resource_kind IS NULL OR v_expected_count NOT BETWEEN 0 AND 250 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile lineage page is stale';
    END IF;
    RETURN QUERY
    WITH resources AS (
        SELECT usr.id, usr.profile_schema_version_id, usr.revision,
               usr.profile_values::text AS values, usr.profile_sha256
        FROM iam_users AS usr WHERE v_resource_kind = 'user'
        UNION ALL
        SELECT organization.id, organization.profile_schema_version_id,
               organization.revision, organization.profile_values::text,
               organization.profile_sha256
        FROM iam_organizations AS organization
        WHERE v_resource_kind = 'organization'
        UNION ALL
        SELECT role.id, role.profile_schema_version_id, role.revision,
               role.profile_values::text, role.profile_sha256
        FROM iam_roles AS role WHERE v_resource_kind = 'role'
    )
    SELECT resource.id, resource.profile_schema_version_id,
           resource.revision, resource.values, resource.profile_sha256,
           false, NULL::bytea
    FROM resources AS resource
    JOIN iam_profile_schema_versions AS actual
      ON actual.id = resource.profile_schema_version_id
     AND actual.resource_kind = v_resource_kind
     AND actual.state IN ('active', 'retired')
     AND actual.version <= v_source_version
    WHERE (v_after_resource_id IS NULL OR resource.id > v_after_resource_id)
      AND resource.id <= v_through_resource_id
    ORDER BY resource.id LIMIT 250;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_lineage_resources_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) FROM PUBLIC;
