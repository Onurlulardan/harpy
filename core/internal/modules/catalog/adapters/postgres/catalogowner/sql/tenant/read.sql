SELECT component.component_id::text, component.contract,
       COALESCE(staged.content_sha256,component.content_sha256),
       CASE WHEN staged.operation_id IS NULL
            THEN component.lifecycle_state ELSE 'STAGED' END
FROM catalog_catalog_components AS component
LEFT JOIN LATERAL (
    SELECT stage.operation_id,stage.content_sha256
    FROM catalog_catalog_component_receipts AS stage
    WHERE stage.component_id=component.component_id
      AND stage.action='stage'
      AND NOT EXISTS (
          SELECT 1 FROM catalog_catalog_component_receipts AS resolved
          WHERE resolved.operation_id=stage.operation_id
            AND resolved.component_id=stage.component_id
            AND resolved.action IN ('cleanup','activate')
      )
    ORDER BY stage.completed_at DESC,stage.receipt_id DESC
    LIMIT 1
) AS staged ON true
WHERE component.component_id=$1::uuid
  AND component.owner_generation=$2::bigint
  AND component.writer_state='v2'
