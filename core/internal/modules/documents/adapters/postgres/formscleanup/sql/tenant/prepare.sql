WITH operation AS MATERIALIZED (
    SELECT * FROM documents_forms_cleanup_operations
    WHERE operation_id = $1::uuid FOR UPDATE
), adoption AS MATERIALIZED (
    SELECT receipt.receipt_sha256
    FROM documents_form_record_adoption_receipts AS receipt
    JOIN operation ON receipt.definition_id = operation.definition_id
      AND receipt.reference_epoch = operation.reference_epoch
    WHERE operation.lifecycle_state = 'REQUESTED'
      AND receipt.lifecycle_state = 'SUCCEEDED'
      AND receipt.receipt_sha256 IS NOT NULL
), matched_links AS MATERIALIZED (
    SELECT link.*
    FROM documents_links AS link
    JOIN operation ON true
    JOIN adoption ON true
    WHERE operation.lifecycle_state = 'REQUESTED'
      AND link.source_owner = 'FORMS'
      AND link.source_kind = 'form_record'
      AND link.source_id = operation.definition_id
      AND link.usage = 'LIFECYCLE'
      AND link.lifecycle_state = 'ACTIVE'
    FOR UPDATE OF link
), link_evidence AS (
    SELECT count(link.id)::bigint AS link_count,
           count(DISTINCT link.item_id)::bigint AS item_count,
           pg_catalog.sha256(pg_catalog.convert_to(
               pg_catalog.encode(adoption.receipt_sha256, 'hex') || '|' ||
               COALESCE(pg_catalog.string_agg(
                   link.id::text || '|' || link.item_id::text || '|' ||
                   link.source_id::text || '|' || link.source_version_id::text || '|' ||
                   link.reference_epoch::text, '|' ORDER BY link.id
               ), ''), 'UTF8'
           )) AS link_catalog_sha256
    FROM adoption LEFT JOIN matched_links AS link ON true
    GROUP BY adoption.receipt_sha256
), inserted_targets AS (
    INSERT INTO documents_forms_cleanup_targets (
        operation_id, item_id, created_at
    )
    SELECT DISTINCT $1::uuid, item_id, $2::timestamptz
    FROM matched_links
    ON CONFLICT (operation_id, item_id) DO NOTHING
    RETURNING item_id
), released_links AS (
    UPDATE documents_links AS link
    SET lifecycle_state = 'RELEASED', released_at = $2::timestamptz
    FROM matched_links AS matched
    WHERE link.id = matched.id
    RETURNING link.id
), started AS (
    UPDATE documents_forms_cleanup_operations AS cleanup
    SET lifecycle_state = 'RUNNING',
        link_count = evidence.link_count,
        item_count = evidence.item_count,
        link_catalog_sha256 = evidence.link_catalog_sha256,
        updated_at = $2::timestamptz
    FROM link_evidence AS evidence
    WHERE cleanup.operation_id = $1::uuid
      AND cleanup.lifecycle_state = 'REQUESTED'
      AND (SELECT count(*) FROM released_links) = evidence.link_count
      AND (SELECT count(*) FROM inserted_targets) <= evidence.item_count
    RETURNING cleanup.operation_id
), eligible AS MATERIALIZED (
    SELECT target.item_id, target.purge_job_id, target.purge_run_at,
           item.lifecycle_state, item.purge_starts_at
    FROM documents_forms_cleanup_targets AS target
    JOIN documents_items AS item ON item.id = target.item_id
    JOIN documents_forms_cleanup_operations AS cleanup
      ON cleanup.operation_id = target.operation_id
    WHERE target.operation_id = $1::uuid
      AND target.enqueued_at IS NULL
      AND cleanup.lifecycle_state = 'RUNNING'
      AND item.lifecycle_state IN (
          'ACTIVE','TRASHED','PURGE_COOLING','PURGING'
      )
      AND NOT EXISTS (
          SELECT 1 FROM documents_legal_holds AS hold
          WHERE hold.item_id = item.id AND hold.lifecycle_state = 'ACTIVE'
      )
      AND NOT EXISTS (
          SELECT 1 FROM documents_links AS blocker
          WHERE blocker.item_id = item.id
            AND blocker.lifecycle_state = 'ACTIVE'
            AND blocker.usage = 'LIFECYCLE'
            AND NOT (
                blocker.source_owner = 'FORMS'
                AND blocker.source_kind = 'form_record'
                AND blocker.source_id = cleanup.definition_id
            )
      )
    FOR UPDATE OF target, item
), cooled AS (
    UPDATE documents_items AS item
    SET lifecycle_state = 'PURGE_COOLING',
        resume_state = item.lifecycle_state,
        purge_requested_at = $2::timestamptz,
        purge_starts_at = $2::timestamptz + interval '1 microsecond',
        purge_reason = 'Forms definition cleanup',
        row_version = item.row_version + 1,
        updated_at = $2::timestamptz
    FROM eligible
    WHERE item.id = eligible.item_id
      AND eligible.lifecycle_state IN ('ACTIVE','TRASHED')
    RETURNING item.id
), candidates AS (
    SELECT eligible.item_id, eligible.purge_job_id,
           CASE
               WHEN eligible.purge_run_at IS NOT NULL
                   THEN eligible.purge_run_at
               WHEN eligible.lifecycle_state IN ('ACTIVE','TRASHED')
                   THEN $2::timestamptz + interval '1 microsecond'
               WHEN eligible.lifecycle_state = 'PURGING'
                   THEN $2::timestamptz
               ELSE eligible.purge_starts_at
           END AS purge_run_at
    FROM eligible
    CROSS JOIN (SELECT count(*) FROM cooled) AS cooled_barrier
)
SELECT cleanup.operation_id, cleanup.definition_id, cleanup.reference_epoch,
       cleanup.lifecycle_state, cleanup.link_count, cleanup.item_count,
       cleanup.version_count, cleanup.ciphertext_bytes,
       cleanup.link_catalog_sha256, cleanup.zero_blocker_sha256,
       cleanup.receipt_sha256, cleanup.requested_at, cleanup.completed_at,
       candidates.item_id, candidates.purge_job_id, candidates.purge_run_at
FROM documents_forms_cleanup_operations AS cleanup
LEFT JOIN candidates ON true
WHERE cleanup.operation_id = $1::uuid
ORDER BY candidates.item_id;
