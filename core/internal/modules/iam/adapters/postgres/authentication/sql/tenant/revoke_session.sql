SELECT outcome, locator_selector, locator_expected_revision,
       locator_revision, locator_source_sha256, locator_state,
       locator_projected_at, locator_retain_until
FROM iam_session_revoke_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7, $8
);
