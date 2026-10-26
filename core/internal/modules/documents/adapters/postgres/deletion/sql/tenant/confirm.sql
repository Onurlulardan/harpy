WITH fence AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'documents:dr-hold:' || pg_catalog.current_schema(), 0
        )
    )
), changed_deletion AS (
    UPDATE documents_blob_deletions d
    SET lifecycle_state = 'CONFIRMED', provider_request_id = $3,
        delete_response_sha256 = $4, absence_probe_sha256 = $5,
        billing_adjustment_receipt_id = $6::uuid, confirmed_at = $7,
        updated_at = $7
    WHERE d.id = $1::uuid AND d.lifecycle_state = 'RECONCILING'
      AND d.lease_generation = $2
      AND EXISTS (SELECT 1 FROM fence)
      AND (d.target_kind = 'UPLOAD_ATTEMPT' OR NOT EXISTS (
          SELECT 1 FROM documents_dr_holds AS hold
          WHERE hold.lifecycle_state IN ('HOLDING','COPYING','VERIFIED')
      ))
    RETURNING d.*
), changed_attempt AS (
    UPDATE documents_upload_attempts a
    SET lifecycle_state = 'CLEANED', cleanup_receipt_sha256 = $5,
        updated_at = $7
    FROM changed_deletion d
    WHERE d.target_kind = 'UPLOAD_ATTEMPT' AND a.upload_id = d.target_id
      AND a.attempt_generation = d.object_generation
      AND a.lifecycle_state = 'CLEANUP_PENDING'
    RETURNING a.upload_id, a.attempt_generation
), changed_upload AS (
    UPDATE documents_uploads u
    SET cleanup_required = false, cleanup_receipt_sha256 = $5,
        updated_at = $7
    FROM changed_attempt a
    WHERE u.id = a.upload_id AND u.attempt_generation = a.attempt_generation
      AND u.cleanup_required
    RETURNING u.id
), changed_version AS (
    UPDATE documents_versions v
    SET lifecycle_state = 'PURGED', delete_receipt_sha256 = $5, purged_at = $7
    FROM changed_deletion d
    WHERE d.target_kind = 'VERSION' AND v.id = d.target_id
      AND v.lifecycle_state = 'DELETE_PENDING'
    RETURNING v.id
), changed_archive AS (
    UPDATE documents_archives a
    SET lifecycle_state = 'PURGED', delete_receipt_sha256 = $5,
        purged_at = $7, updated_at = $7
    FROM changed_deletion d
    WHERE d.target_kind = 'ARCHIVE' AND a.id = d.target_id
      AND a.lifecycle_state = 'EXPIRED'
    RETURNING a.id
), target_receipt AS (
    SELECT id FROM changed_upload
    UNION ALL SELECT id FROM changed_version
    UNION ALL SELECT id FROM changed_archive
)
SELECT d.confirmed_at
FROM changed_deletion d JOIN target_receipt t ON t.id = d.target_id;
