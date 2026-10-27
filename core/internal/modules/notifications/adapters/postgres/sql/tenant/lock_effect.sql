SELECT pg_catalog.pg_advisory_xact_lock(
    pg_catalog.hashtextextended($1 || ':' || pg_catalog.encode($2, 'hex') || ':' || $3::text, 0)
)
WHERE (
    SELECT pg_catalog.count(*)
    FROM notifications_effect_receipts
    WHERE FALSE
) = 0;
