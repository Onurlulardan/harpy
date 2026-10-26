SELECT id, environment_id, name, description, enabled, service_name,
       condition, cooldown_minutes, channels, created_at, updated_at
FROM platform.platformops_get_alert_rule_v1($1::uuid,$2::uuid)
