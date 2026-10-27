UPDATE catalog_installations
SET active_version = $3::text,
    state = $4::text,
    reference_epoch = $5::bigint,
    row_version = $6::bigint,
    updated_at = $7::timestamptz
WHERE installation_id = $1::uuid AND package_id = $2::uuid AND row_version = $8::bigint
RETURNING installation_id, package_id, active_version, state, reference_epoch, row_version,
    (
        SELECT operation_id
        FROM catalog_operation_runs
        WHERE installation_id = catalog_installations.installation_id
        ORDER BY started_at DESC, operation_id DESC
        LIMIT 1
    )
