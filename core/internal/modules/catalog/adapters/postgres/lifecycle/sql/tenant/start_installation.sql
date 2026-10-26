INSERT INTO catalog_installations (
    installation_id, package_id, active_version, state,
    reference_epoch, row_version, updated_at
) VALUES ($1::uuid, $2::uuid, $3::text, $4::text, $5::bigint, $6::bigint, $7::timestamptz)
ON CONFLICT (package_id) DO UPDATE SET
    active_version = EXCLUDED.active_version,
    state = EXCLUDED.state,
    reference_epoch = EXCLUDED.reference_epoch,
    row_version = EXCLUDED.row_version,
    updated_at = EXCLUDED.updated_at
WHERE $8::text IS NOT NULL
  AND catalog_installations.installation_id = EXCLUDED.installation_id
  AND catalog_installations.state = $8::text
  AND catalog_installations.row_version = $9::bigint
RETURNING installation_id
