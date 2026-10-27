WITH inserted AS (
    INSERT INTO documents_reverse_reconciliations (
        reconciliation_id, source_generation, target_generation,
        source_count, target_count, source_sha256, target_sha256,
        lifecycle_state, requested_by, created_at
    )
    VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, 'PENDING', $8::uuid, $9)
    ON CONFLICT (reconciliation_id) DO NOTHING
    RETURNING reconciliation_id
)
SELECT true FROM inserted
UNION ALL
SELECT true FROM documents_reverse_reconciliations r
WHERE r.reconciliation_id = $1::uuid AND NOT EXISTS (SELECT 1 FROM inserted)
  AND r.source_generation = $2 AND r.target_generation = $3
  AND r.source_count = $4 AND r.target_count = $5
  AND r.source_sha256 = $6 AND r.target_sha256 = $7
  AND r.requested_by = $8::uuid;
