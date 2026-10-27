SELECT row_index,outcome,owner_kind,owner_id,owner_revision,error_code,
       source_sha256,result_sha256
FROM iam_system_record_import_worker_receipts_v1(
    $1::name,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
)
