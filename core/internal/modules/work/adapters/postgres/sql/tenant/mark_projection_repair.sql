UPDATE work_projection_streams
SET state = 'GAP', state_sha256 = $5, updated_at = transaction_timestamp()
WHERE source_owner = $1 AND aggregate_kind = $2 AND aggregate_id = $3
  AND through_sequence < $4
