SELECT session_revision, family_revision, token_revision, issued_at,
       idle_expires_at, absolute_expires_at, retain_until,
       platform_authorization_clock, security_epoch
FROM iam_session_issue_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
);
