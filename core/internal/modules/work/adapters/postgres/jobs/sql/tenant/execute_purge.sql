WITH eligible AS (
    SELECT operation_id, target_id, expected_reference_epoch, blocker_sha256
    FROM work_purge_operations
    WHERE operation_id = $1 AND target_kind = 'CALENDAR_EVENT'
      AND state IN ('COOLING','EXECUTING') AND cooling_until <= $2
      AND NOT legal_hold AND blocker_count = 0
    FOR UPDATE
), purged AS (
    UPDATE work_calendar_events event
    SET title = '[purged]', description = '', generated_state = 'CONTENT_PURGED',
        row_version = row_version + 1, updated_at = $2
    FROM eligible
    WHERE event.id = eligible.target_id AND event.generated
      AND event.generated_state = 'ACTIVE' AND event.live_reference_count = 0
      AND NOT event.legal_hold
    RETURNING eligible.operation_id, eligible.target_id,
        eligible.expected_reference_epoch, eligible.blocker_sha256
), tombstone AS (
    INSERT INTO work_id_tombstones (
        target_kind, target_id, reference_epoch, receipt_sha256, purged_at, retain_until
    )
    SELECT 'CALENDAR_EVENT', target_id, expected_reference_epoch,
        blocker_sha256, $2, $2 + interval '365 days'
    FROM purged
    ON CONFLICT (target_kind, target_id) DO NOTHING
    RETURNING target_id
)
UPDATE work_purge_operations operation
SET state = 'SUCCEEDED', destructive_started_at = COALESCE(destructive_started_at, $2),
    completed_at = $2, updated_at = $2
WHERE operation.operation_id = $1
  AND EXISTS (SELECT 1 FROM tombstone)
