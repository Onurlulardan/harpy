-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_publication_receipt_read_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_manage_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    command_key text, request_sha256 bytea, result_id uuid,
    target_schema_id uuid, result_revision bigint, authorization_clock bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR NOT iam_profile_actor_assert_v1(
           p_scope_schema, p_actor_id, p_manage_permission_id,
           p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile publication receipt authority is denied';
    END IF;
    RETURN QUERY
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           plan.target_schema_id, receipt.result_revision, receipt.clock_value
    FROM iam_authorization_operation_receipts AS receipt
    LEFT JOIN iam_profile_backfill_plans AS plan
      ON plan.id = receipt.result_id
    WHERE receipt.operation_id = p_operation_id
      AND receipt.command_key IN (
          'iam.profile_schema.publish', 'iam.profile_backfill.plan_create'
      );
END;
$function$;
CREATE FUNCTION iam_profile_backfill_source_resources_plan_v1(
    p_scope_schema name, p_resource_kind text, p_source_schema_id uuid,
    p_after_resource_id uuid, p_limit integer, p_actor_id uuid,
    p_manage_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    resource_id uuid, source_schema_version_id uuid, source_revision bigint,
    profile_values text, profile_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_source_schema_id IS NULL OR p_limit NOT BETWEEN 1 AND 250 OR
       (p_after_resource_id IS NOT NULL AND p_after_resource_id =
            '00000000-0000-0000-0000-000000000000'::uuid) OR
       NOT EXISTS (
           SELECT 1 FROM iam_profile_schema_versions AS schema_version
           WHERE schema_version.id = p_source_schema_id
             AND schema_version.resource_kind = p_resource_kind
             AND schema_version.state = 'active'
       ) OR NOT iam_profile_actor_assert_v1(
           p_scope_schema, p_actor_id, p_manage_permission_id,
           p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile backfill planning authority is denied';
    END IF;
    IF p_resource_kind = 'user' THEN
        LOCK TABLE iam_users IN SHARE MODE;
        RETURN QUERY
        SELECT resource.id, resource.profile_schema_version_id,
               resource.revision, resource.profile_values::text,
               resource.profile_sha256
        FROM iam_users AS resource
        JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.profile_schema_version_id
         AND row_schema.resource_kind = p_resource_kind
         AND row_schema.state IN ('active', 'retired')
        JOIN iam_profile_schema_versions AS source_schema
          ON source_schema.id = p_source_schema_id
         AND row_schema.version <= source_schema.version
        WHERE p_after_resource_id IS NULL OR resource.id > p_after_resource_id
        ORDER BY resource.id LIMIT p_limit;
    ELSIF p_resource_kind = 'organization' THEN
        LOCK TABLE iam_organizations IN SHARE MODE;
        RETURN QUERY
        SELECT resource.id, resource.profile_schema_version_id,
               resource.revision, resource.profile_values::text,
               resource.profile_sha256
        FROM iam_organizations AS resource
        JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.profile_schema_version_id
         AND row_schema.resource_kind = p_resource_kind
         AND row_schema.state IN ('active', 'retired')
        JOIN iam_profile_schema_versions AS source_schema
          ON source_schema.id = p_source_schema_id
         AND row_schema.version <= source_schema.version
        WHERE p_after_resource_id IS NULL OR resource.id > p_after_resource_id
        ORDER BY resource.id LIMIT p_limit;
    ELSE
        LOCK TABLE iam_roles IN SHARE MODE;
        RETURN QUERY
        SELECT resource.id, resource.profile_schema_version_id,
               resource.revision, resource.profile_values::text,
               resource.profile_sha256
        FROM iam_roles AS resource
        JOIN iam_profile_schema_versions AS row_schema
          ON row_schema.id = resource.profile_schema_version_id
         AND row_schema.resource_kind = p_resource_kind
         AND row_schema.state IN ('active', 'retired')
        JOIN iam_profile_schema_versions AS source_schema
          ON source_schema.id = p_source_schema_id
         AND row_schema.version <= source_schema.version
        WHERE p_after_resource_id IS NULL OR resource.id > p_after_resource_id
        ORDER BY resource.id LIMIT p_limit;
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_publication_receipt_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_profile_backfill_source_resources_plan_v1(
    name,text,uuid,uuid,integer,uuid,uuid,bigint,bigint
) FROM PUBLIC;
