SELECT role.id,role.scope_kind,role.scope_id
FROM iam_roles AS role
WHERE role.id=$1 AND role.status='active' AND NOT role.is_protected
  AND ((role.scope_kind='tenant' AND role.scope_id=$2) OR
       (role.scope_kind='organization' AND EXISTS(
           SELECT 1 FROM iam_organizations AS organization
           WHERE organization.id=role.scope_id AND organization.status='active'
       )))
LIMIT 1
FOR KEY SHARE
