SELECT import_id,job_id,resource_kind,state,processed_rows,created_rows,
       updated_rows,skipped_rows,rejected_rows,created_at,completed_at
FROM iam_system_record_import_cancel_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6
)
