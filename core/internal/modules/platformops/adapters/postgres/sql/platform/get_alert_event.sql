SELECT id, environment_id, alert_rule_id, service_name, previous_status,
       current_status, channels_used, message, resolved_at, created_at
FROM platform.platformops_get_alert_event_v1($1::uuid,$2::uuid)
