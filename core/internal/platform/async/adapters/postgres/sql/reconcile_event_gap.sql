SELECT completed_at, requeued, action_id, disposition
FROM platform.platform_async_reconcile_event_gap_v2($1, $2, $3, $4);
