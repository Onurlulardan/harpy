WITH observed AS (
    SELECT max(event_sequence) AS through_sequence
    FROM (
        SELECT event_sequence FROM work_inbox_items WHERE source_id = $1
        UNION ALL
        SELECT event_sequence FROM work_calendar_items WHERE source_id = $1
    ) item
)
UPDATE work_projection_streams stream
SET through_sequence = observed.through_sequence,
    bootstrap_sequence = greatest(
        stream.bootstrap_sequence, observed.through_sequence),
    state = 'CURRENT', updated_at = $2
FROM observed
WHERE stream.aggregate_id = $1 AND stream.state IN ('GAP','REBUILDING')
  AND observed.through_sequence IS NOT NULL
  AND observed.through_sequence >= stream.through_sequence
