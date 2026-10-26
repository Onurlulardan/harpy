SELECT token_version, authorization_version
FROM iam_rest_query_actor_versions_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid
)
