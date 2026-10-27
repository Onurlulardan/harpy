WITH changed AS (
    UPDATE documents_reverse_reconciliations
    SET lifecycle_state = CASE WHEN $3::boolean THEN 'ACCEPTED' ELSE 'REJECTED' END,
        approved_by = $2::uuid, decided_at = $4
    WHERE reconciliation_id = $1::uuid AND lifecycle_state = 'PENDING'
      AND requested_by IS DISTINCT FROM $2::uuid
    RETURNING reconciliation_id
)
SELECT true FROM changed
UNION ALL
SELECT true FROM documents_reverse_reconciliations r
WHERE r.reconciliation_id = $1::uuid AND NOT EXISTS (SELECT 1 FROM changed)
  AND r.approved_by = $2::uuid AND r.lifecycle_state =
      CASE WHEN $3::boolean THEN 'ACCEPTED' ELSE 'REJECTED' END;
