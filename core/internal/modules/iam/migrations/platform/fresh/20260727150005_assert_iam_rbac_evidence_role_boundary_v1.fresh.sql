-- +goose Up
SELECT 1 / CASE WHEN (
    SELECT NOT role.rolsuper
       AND NOT role.rolinherit
       AND NOT role.rolcreaterole
       AND NOT role.rolcreatedb
       AND NOT role.rolcanlogin
       AND NOT role.rolbypassrls
    FROM pg_catalog.pg_roles AS role
    WHERE role.rolname = 'forgecrud_owner'
) AND (
    SELECT NOT role.rolsuper
       AND NOT role.rolinherit
       AND NOT role.rolcreaterole
       AND NOT role.rolcreatedb
       AND NOT role.rolcanlogin
       AND NOT role.rolbypassrls
    FROM pg_catalog.pg_roles AS role
    WHERE role.rolname = 'forgecrud_runtime'
) AND (
    SELECT NOT role.rolsuper
       AND NOT role.rolinherit
       AND NOT role.rolcreaterole
       AND NOT role.rolcreatedb
       AND role.rolcanlogin
       AND NOT role.rolbypassrls
    FROM pg_catalog.pg_roles AS role
    WHERE role.rolname = 'forgecrud_v2_runtime'
) AND EXISTS (
    SELECT 1
    FROM pg_catalog.pg_auth_members AS membership
    WHERE membership.roleid = 'forgecrud_runtime'::pg_catalog.regrole
      AND membership.member = 'forgecrud_v2_runtime'::pg_catalog.regrole
      AND NOT membership.inherit_option
      AND membership.set_option
      AND NOT membership.admin_option
) THEN 1 ELSE 0 END AS iam_rbac_evidence_role_boundary_assertion;
