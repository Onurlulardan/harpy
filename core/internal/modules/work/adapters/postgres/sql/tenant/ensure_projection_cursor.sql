INSERT INTO work_projection_streams (
    source_owner, aggregate_kind, aggregate_id, through_sequence,
    bootstrap_sequence, state, state_sha256, updated_at
) VALUES ($1, $2, $3, 0, 0, 'CURRENT', $4, transaction_timestamp())
ON CONFLICT (source_owner, aggregate_kind, aggregate_id) DO NOTHING
