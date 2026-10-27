SELECT export_id,job_id,resource_kind,format,state,processed_rows,
       plaintext_bytes,created_at,ready_at,expires_at,replayed
FROM iam_system_record_export_submit_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
    $12,$13,$14,$15,$16
)
