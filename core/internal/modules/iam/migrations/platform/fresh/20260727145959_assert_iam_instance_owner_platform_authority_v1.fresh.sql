-- +goose Up
SELECT 1 / CASE WHEN (
    SELECT pg_catalog.count(*)
    FROM platform.iam_platform_permissions AS permission
    WHERE permission.permission_key = 'environment_all'
      AND permission.match_kind = 'environment_all'
      AND permission.deployment_mode = 'both'
      AND permission.owner_only
      AND NOT permission.delegable
      AND NOT permission.step_up_required
      AND permission.state = 'active'
) = 1 THEN 1 ELSE 0 END AS iam_instance_owner_platform_authority_assertion;
