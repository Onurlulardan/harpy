SELECT state,processed_rows,created_rows,updated_rows,skipped_rows,
       rejected_rows,decompressed_bytes
FROM iam_system_record_import_worker_finish_v1(
    $1::name,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
)
