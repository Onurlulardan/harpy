SELECT extension_row.extversion,
       namespace_row.nspname,
       owner_role.rolname
FROM pg_catalog.pg_extension AS extension_row
JOIN pg_catalog.pg_namespace AS namespace_row
  ON namespace_row.oid = extension_row.extnamespace
JOIN pg_catalog.pg_roles AS owner_role ON owner_role.oid = extension_row.extowner
WHERE extension_row.extname = $1;
