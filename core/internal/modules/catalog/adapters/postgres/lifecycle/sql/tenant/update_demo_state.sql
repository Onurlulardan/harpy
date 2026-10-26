UPDATE catalog_demo_states
SET source_run_id = $3::uuid, state = $4::text, record_ids = $5::jsonb,
    last_receipt_id = $6::uuid, revision = revision + 1,
    updated_at = $7::timestamptz
WHERE installation_id = $1::uuid AND component_id = $2::uuid
  AND content_sha256 = $8::text AND revision = $9::bigint
RETURNING revision

