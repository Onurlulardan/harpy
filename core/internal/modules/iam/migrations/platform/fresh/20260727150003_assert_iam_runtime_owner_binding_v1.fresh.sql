-- +goose Up
SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM pg_catalog.pg_auth_members AS membership
    WHERE membership.roleid = 'harpy_runtime'::pg_catalog.regrole
      AND membership.member = 'harpy_v2_runtime'::pg_catalog.regrole
      AND NOT membership.inherit_option
      AND membership.set_option
) THEN 1 ELSE 0 END AS iam_runtime_owner_binding_assertion;
