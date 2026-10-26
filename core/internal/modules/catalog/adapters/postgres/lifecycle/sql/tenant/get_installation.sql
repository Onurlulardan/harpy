SELECT i.installation_id, i.package_id, i.active_version, i.state,
       i.reference_epoch, i.row_version, last_operation.operation_id
FROM catalog_installations AS i
LEFT JOIN LATERAL (
    SELECT operation_id
    FROM catalog_operation_runs
    WHERE installation_id = i.installation_id
    ORDER BY started_at DESC, operation_id DESC
    LIMIT 1
) AS last_operation ON true
WHERE i.package_id = $1::uuid
