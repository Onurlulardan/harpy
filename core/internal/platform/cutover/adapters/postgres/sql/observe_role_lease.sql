SELECT pg_catalog.count(*) = 1
FROM pg_catalog.pg_roles AS role
WHERE role.rolname = $1
  AND role.rolcanlogin
  AND NOT role.rolinherit
  AND NOT role.rolsuper
  AND NOT role.rolcreatedb
  AND NOT role.rolcreaterole
  AND role.rolreplication = $3
  AND NOT role.rolbypassrls
  AND role.rolconnlimit = $4
  AND role.rolvaliduntil IS NOT DISTINCT FROM $2::timestamptz
  AND role.rolvaliduntil > pg_catalog.clock_timestamp()
  AND NOT EXISTS (
      SELECT 1 FROM pg_catalog.pg_auth_members AS membership
      WHERE membership.roleid = role.oid OR membership.member = role.oid
  )
