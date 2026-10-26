SELECT EXISTS (
    SELECT 1 FROM platform.notification_realtime_tickets
    WHERE tenant_id = $1::uuid
      AND COALESCE(consumed_at, expires_at) <= $2::timestamptz - interval '24 hours'
) OR EXISTS (
    SELECT 1 FROM platform.notification_provider_locators
    WHERE tenant_id = $1::uuid AND expires_at <= $2::timestamptz
) OR EXISTS (
    SELECT 1 FROM platform.notification_realtime_ticket_audits
    WHERE tenant_id = $1::uuid AND retain_until <= $2::timestamptz
) OR EXISTS (
    SELECT 1 FROM platform.notification_legacy_connection_hits
    WHERE tenant_id = $1::uuid
      AND occurred_at <= $2::timestamptz - interval '90 days'
);
