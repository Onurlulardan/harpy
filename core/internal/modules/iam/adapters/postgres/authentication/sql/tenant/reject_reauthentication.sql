SELECT state, revision, attempt_count
FROM iam_reauthentication_reject_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7
);
