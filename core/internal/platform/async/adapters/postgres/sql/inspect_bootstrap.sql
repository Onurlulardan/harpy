WITH stream AS MATERIALIZED (
    SELECT
        pg_catalog.bool_or(
            row.id = $9::uuid AND row.record_kind = 'watermark' AND
            row.watermark_kind = 'bootstrap' AND row.tenant_id = $10::uuid AND
            row.producer_group_id = $11::text AND
            row.producer_group_generation = $12::bigint AND
            row.producer_domain = $13::text AND row.producer_generation = $14::bigint AND
            row.writer_origin = $15::text AND row.consumer_group_id = $3::text AND
            row.consumer_group_generation = $4::bigint AND
            row.consumer_domain = $5::text AND row.consumer_generation = $6::bigint AND
            row.aggregate_type = $16::text AND row.aggregate_id = $17::uuid AND
            row.through_sequence = $18::bigint AND row.causation_id = $19::uuid AND
            row.state_sha256 = $20::text AND row.receipt_sha256 = $21::text
        ) AS duplicate,
        pg_catalog.count(row.id) AS row_count
    FROM platform.inbox AS row
    WHERE row.consumer_name = $7::text AND row.consumer_version = $8::integer
      AND row.scope_kind = $1::text AND row.scope_id = $2::uuid
      AND row.aggregate_type = $16::text AND row.aggregate_id = $17::uuid
)
SELECT CASE
    WHEN $10::uuid <> $2::uuid OR $15::text <> 'v2' THEN 'invalid_scope'
    WHEN COALESCE(stream.duplicate, false) THEN 'duplicate'
    WHEN stream.row_count > 0 THEN 'stream_not_empty'
    ELSE 'apply'
END
FROM stream;
