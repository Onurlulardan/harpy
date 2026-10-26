SELECT operation_id,tenant_id,schema_name,user_id,owner_binding_id,request_sha256
FROM platform.iam_installation_instance_owner_input_v1($1::uuid);
