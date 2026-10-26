SELECT row_index,outcome,COALESCE(owner_kind,''),owner_id,
       COALESCE(error_code,''),source_sha256,result_sha256
FROM iam_system_record_import_receipts_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7
)
