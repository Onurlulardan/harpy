INSERT INTO work_aggregate_sequences (
    aggregate_kind, aggregate_id, event_sequence, aggregate_version, updated_at
) VALUES ($1, $2, 1, $3, transaction_timestamp())
ON CONFLICT (aggregate_kind, aggregate_id) DO UPDATE
SET event_sequence = work_aggregate_sequences.event_sequence + 1,
    aggregate_version = EXCLUDED.aggregate_version,
    updated_at = transaction_timestamp()
WHERE work_aggregate_sequences.aggregate_version < EXCLUDED.aggregate_version
RETURNING event_sequence
