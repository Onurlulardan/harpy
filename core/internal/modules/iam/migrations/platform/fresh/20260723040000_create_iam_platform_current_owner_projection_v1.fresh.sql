-- +goose Up
-- Migration stream version: 20260723040000.
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_current_owner_projection_v1(
    p_expected_environment_id uuid
)
RETURNS TABLE(
    environment_id uuid,
    owner_binding_id uuid,
    owner_policy_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT policy.environment_id,policy.owner_binding_id,policy.revision
      FROM platform.iam_platform_owner_policies AS policy
      JOIN platform.iam_platform_subject_roles AS binding
        ON binding.environment_id = policy.environment_id
       AND binding.id = policy.owner_binding_id
       AND binding.state = 'active'
      JOIN platform.iam_platform_roles AS role
        ON role.environment_id = binding.environment_id
       AND role.id = binding.role_id
       AND role.role_kind = 'instance_owner'
       AND role.state = 'active'
      JOIN platform.iam_platform_subjects AS subject
        ON subject.environment_id = binding.environment_id
       AND subject.id = binding.subject_id
       AND subject.state = 'active'
     WHERE policy.environment_id = p_expected_environment_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_platform_current_owner_projection_v1(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_platform_current_owner_projection_v1(uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_platform_current_owner_projection_v1(uuid)
TO forgecrud_restore;
