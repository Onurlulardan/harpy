INSERT INTO work_audit_entries (
    audit_id, operation_id, actor_id, action, aggregate_kind, aggregate_id,
    aggregate_version, event_sequence, reason, occurred_at
) VALUES ($1, $2, $3, $4, $4, $5, $6, $7, NULL, transaction_timestamp())
