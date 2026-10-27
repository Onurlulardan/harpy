SELECT state, policy_attempts, revision
FROM iam_initial_password_reject_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5
)

