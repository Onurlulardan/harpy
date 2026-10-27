SELECT outcome, issued_at, locator_selector, locator_expected_revision,
       locator_revision, locator_source_sha256, locator_state,
       locator_projected_at, locator_retain_until, successor_selector,
       successor_revision, successor_source_sha256, successor_state,
       successor_issued_at, successor_projected_at,
       successor_family_expires_at, successor_retain_until
FROM iam_refresh_rotate_v4(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
);
