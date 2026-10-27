SELECT COALESCE(pg_catalog.max(result_settings_revision), 0)::bigint
FROM notifications_smtp_settings_receipts;
