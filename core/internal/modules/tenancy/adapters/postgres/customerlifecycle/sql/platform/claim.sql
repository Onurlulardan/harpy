SELECT operation_id,control_tenant_id,target_tenant_id,kind,state,
       result_revision,false,created_at,updated_at,completed_at,failure_code,
       schema_name,retention_receipt_id,backup_receipt_id,
       contributor_receipt_id,evidence_sha256,proof_verified_at
FROM platform.tenancy_customer_lifecycle_claim_v1($1::uuid,$2::timestamptz,$3::integer)
