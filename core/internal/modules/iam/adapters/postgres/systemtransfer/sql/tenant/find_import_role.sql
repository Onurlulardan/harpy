SELECT role.id,role.revision
FROM iam_roles AS role
WHERE role.scope_kind=$1
  AND role.scope_id=$2
  AND role.name_key=pg_catalog.convert_to($3,'UTF8')
LIMIT 1
FOR KEY SHARE
