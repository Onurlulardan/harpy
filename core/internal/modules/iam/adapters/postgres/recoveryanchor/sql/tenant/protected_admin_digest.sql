SELECT protected_admin_binding_set_sha256
FROM iam_installation_identity_facts
WHERE environment_id=$1::uuid AND tenant_id=$2::uuid;
