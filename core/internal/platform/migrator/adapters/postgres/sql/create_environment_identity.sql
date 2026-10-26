INSERT INTO platform.environment_identity (
    singleton,
    environment_id,
    database_origin,
    database_system_identifier,
    database_oid,
    release_id,
    bootstrap_authority_sha256
) VALUES (
    TRUE,
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
ON CONFLICT (singleton) DO NOTHING;
