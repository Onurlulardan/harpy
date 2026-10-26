SELECT import_id,job_id,resource_kind,state,processed_rows,created_rows,
       updated_rows,skipped_rows,rejected_rows,created_at,completed_at,replayed
FROM iam_system_record_import_submit_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
    $12,$13,$14,$15,$16,$17,$18,$19
)
