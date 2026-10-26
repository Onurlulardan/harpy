SELECT public_id, tenant_id, release_id, resource_kind, resource_version_id,
       audience, state, generation, capability_sha256, capability_expires_at
FROM platform.analytics_public_resolve_v1($1::bytea);
