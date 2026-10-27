SELECT
    ordinal,
    generation,
    attempt,
    operation_id,
    transaction_class,
    executor_kind,
    identifier_mapping::text,
    arguments::text,
    step_sha256,
    checkpoint_record_id::text,
    processed_rows
FROM platform.schema_job_steps
WHERE job_id = $1
ORDER BY ordinal;
