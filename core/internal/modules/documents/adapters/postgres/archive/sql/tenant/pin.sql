WITH candidate AS MATERIALIZED (
    SELECT a.* FROM documents_archives a
    WHERE a.id = $1::uuid AND a.lifecycle_state = 'PENDING'
      AND a.entry_count = 0 AND a.object_version IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM documents_restore_transient_dispositions AS disposition
          WHERE disposition.target_kind = 'ARCHIVE'
            AND disposition.target_id = a.id
      )
    FOR UPDATE
), requested AS MATERIALIZED (
    SELECT value::uuid AS item_id, (ordinality - 1)::integer AS entry_order
    FROM candidate c,
    LATERAL pg_catalog.jsonb_array_elements_text(
        (pg_catalog.convert_from(c.request_document, 'UTF8')::jsonb)->'item_ids'
    ) WITH ORDINALITY AS request(value, ordinality)
), selected AS MATERIALIZED (
    SELECT r.entry_order, i.id AS document_id, v.id AS version_id,
           v.plaintext_bytes, v.plaintext_sha256
    FROM requested r JOIN documents_items i ON i.id = r.item_id
    JOIN documents_versions v ON v.id = i.current_version_id
    WHERE i.lifecycle_state IN ('ACTIVE','TRASHED','PURGE_COOLING')
      AND v.lifecycle_state = 'READY'
), accepted AS MATERIALIZED (
    SELECT count(*) AS entry_count, pg_catalog.sum(plaintext_bytes) AS total_bytes,
           pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.string_agg(
               document_id::text || ':' || version_id::text || ':' ||
               plaintext_bytes::text || ':' || pg_catalog.encode(plaintext_sha256, 'hex'),
               '|' ORDER BY entry_order), 'UTF8')) AS entries_sha256
    FROM selected
    HAVING count(*) = (SELECT count(*) FROM requested)
       AND count(*) BETWEEN 1 AND 10000
       AND pg_catalog.sum(plaintext_bytes) <= 10737418240
), inserted AS (
    INSERT INTO documents_archive_entries (
        archive_id, entry_order, document_id, version_id,
        plaintext_bytes, plaintext_sha256, added_at
    )
    SELECT $1::uuid, s.entry_order, s.document_id, s.version_id,
           s.plaintext_bytes, s.plaintext_sha256, $2
    FROM selected s CROSS JOIN accepted
    ORDER BY s.entry_order
    RETURNING archive_id
), changed AS (
    UPDATE documents_archives a
    SET entry_count = accepted.entry_count,
        plaintext_bytes = accepted.total_bytes,
        entries_sha256 = accepted.entries_sha256, updated_at = $2
    FROM accepted
    WHERE a.id = $1::uuid
      AND (SELECT count(*) FROM inserted) = accepted.entry_count
    RETURNING a.id
)
SELECT id FROM changed;
