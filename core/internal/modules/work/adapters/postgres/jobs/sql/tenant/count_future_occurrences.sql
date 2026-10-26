SELECT count(*)
FROM work_calendar_events
WHERE generated AND generated_state = 'ACTIVE'
  AND occurrence_start >= transaction_timestamp()
