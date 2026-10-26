SELECT singleton, true
FROM notifications_smtp_settings_lock
WHERE singleton = true AND pg_catalog.octet_length($1::text) = 36
FOR UPDATE;
