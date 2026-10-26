SELECT id, environment_id, name, description, enabled, service_name,
       condition, cooldown_minutes, channels, created_at, updated_at
FROM platform.platformops_update_alert_rule_v1(
    $1::uuid,$2::uuid,$3::text,$4::text,$5::boolean,$6::text,$7::text,
    $8::integer,$9::jsonb,$10::timestamptz
)
