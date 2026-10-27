INSERT INTO work_reverse_reconciliations (
    reconciliation_id, source_generation, target_generation, source_count, target_count,
    source_sha256, target_sha256, state, requested_by, approved_by, created_at, decided_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, 'ACCEPTED', $8, $9,
          transaction_timestamp(), transaction_timestamp())
ON CONFLICT (reconciliation_id) DO NOTHING
