SELECT pg_catalog.to_regclass(
    pg_catalog.current_schema() || '.notifications'
) IS NOT NULL
AND (
    SELECT pg_catalog.count(*) >= 0
    FROM notifications_messages
);
