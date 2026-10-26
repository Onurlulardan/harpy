WITH ranked AS (
    SELECT id, tenant_id, user_id, type, level, title, message, action,
           entity_id, entity, data, channel, is_read, created_at, read_at,
           pg_catalog.row_number() OVER (
               PARTITION BY user_id ORDER BY created_at NULLS FIRST, id
           )::bigint AS inbox_revision
    FROM notifications
)
SELECT id, tenant_id, user_id, type, level, title, message, action,
       entity_id, entity, data, channel, is_read, created_at, read_at,
       inbox_revision
FROM ranked
WHERE ($1::uuid IS NULL OR id > $1::uuid)
ORDER BY id
LIMIT $2;
