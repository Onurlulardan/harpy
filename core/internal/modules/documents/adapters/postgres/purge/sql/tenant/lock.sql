WITH item AS MATERIALIZED (
    SELECT i.* FROM documents_items i WHERE i.id = $1::uuid FOR UPDATE
), flags AS (
    SELECT EXISTS (
               SELECT 1 FROM documents_legal_holds h
               WHERE h.item_id = item.id AND h.lifecycle_state = 'ACTIVE'
           ) AS held,
           EXISTS (
               SELECT 1 FROM documents_links l
               WHERE l.item_id = item.id AND l.lifecycle_state = 'ACTIVE'
                 AND l.usage = 'LIFECYCLE'
           ) AS referenced
    FROM item
)
SELECT item.id, item.owner_id, item.lifecycle_state,
       ((item.lifecycle_state = 'PURGE_COOLING' AND item.purge_starts_at <= $2)
        OR (item.lifecycle_state = 'TRASHED' AND item.retention_ends_at <= $2)
        OR item.lifecycle_state = 'PURGING') AS eligible,
       flags.held, flags.referenced,
       v.id, v.object_key, v.object_version, v.object_generation,
       v.plaintext_bytes, v.ciphertext_bytes
FROM item CROSS JOIN flags
LEFT JOIN documents_versions v ON v.item_id = item.id
ORDER BY v.version_number;
