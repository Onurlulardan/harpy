SELECT id, environment_id, name, description, enabled, service_name,
       condition, cooldown_minutes, channels, created_at, updated_at
FROM platform.platformops_list_alert_rules_v1($1::uuid)
