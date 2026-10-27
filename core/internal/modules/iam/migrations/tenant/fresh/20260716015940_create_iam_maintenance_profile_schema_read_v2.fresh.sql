-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_schema_read_v2(
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
    v_source_schema_id uuid;
    v_source_version bigint;
    v_target_schema_id uuid;
    v_after_resource_id uuid;
    v_through_resource_id uuid;
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
            MESSAGE = 'IAM profile schema snapshot authority is invalid';
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
    WHERE plan.id = p_plan_id AND page.job_id = p_job_id
      AND page.payload_reference = p_payload_reference
      AND page.payload_sha256 = p_payload_sha256 AND page.state = 'queued'
      AND plan.state IN ('planned', 'running');
    IF v_resource_kind IS NULL OR (
        SELECT pg_catalog.count(*)
        FROM iam_profile_field_revisions AS snapshot
        WHERE snapshot.schema_version_id = p_schema_id
          AND snapshot.resource_kind = v_resource_kind
    ) > 250 THEN
        RAISE EXCEPTION USING ERRCODE = '54000',
            MESSAGE = 'IAM profile schema snapshot bound is invalid';
    END IF;
    RETURN QUERY
    SELECT version.id, version.resource_kind, version.version, version.state,
           version.definition_sha256, snapshot.field_id, snapshot.field_key,
           snapshot.value_kind, snapshot.cardinality, snapshot.required,
           snapshot.classification, snapshot.normalization_profile,
           snapshot.field_state, snapshot.validation_schema::text,
           snapshot.validation_sha256
    FROM iam_profile_schema_versions AS version
    LEFT JOIN iam_profile_field_revisions AS snapshot
      ON snapshot.schema_version_id = version.id
     AND snapshot.resource_kind = version.resource_kind
    WHERE version.id = p_schema_id AND version.resource_kind = v_resource_kind
      AND ((version.id = v_source_schema_id AND version.state = 'active') OR
           (version.id = v_target_schema_id AND version.state = 'draft') OR
           (version.state IN ('active','retired') AND
            version.version <= v_source_version AND EXISTS (
                SELECT 1 FROM (
                    SELECT usr.id, usr.profile_schema_version_id
                    FROM iam_users AS usr WHERE v_resource_kind = 'user'
                    UNION ALL
                    SELECT organization.id,
                           organization.profile_schema_version_id
                    FROM iam_organizations AS organization
                    WHERE v_resource_kind = 'organization'
                    UNION ALL
                    SELECT role.id, role.profile_schema_version_id
                    FROM iam_roles AS role WHERE v_resource_kind = 'role'
                ) AS resource
                WHERE resource.profile_schema_version_id = version.id
                  AND (v_after_resource_id IS NULL OR
                       resource.id > v_after_resource_id)
                  AND resource.id <= v_through_resource_id
            )))
    ORDER BY snapshot.field_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_schema_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,uuid
) FROM PUBLIC;
