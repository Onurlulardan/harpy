SELECT environment_id::text,
       database_origin,
       database_system_identifier,
       database_oid,
       release_id::text,
       bootstrap_authority_sha256
FROM platform.environment_identity
WHERE singleton = TRUE;
