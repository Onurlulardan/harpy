SELECT tenant_identity, method, path, slug, tenant_id, legacy_endpoint_id,
       public_id, release_id, state, hit_count, last_hit_at,
       consumers_acked_at, remove_after
FROM platform.analytics_legacy_public_resolve_v1(
    $1::text, $2::text, $3::text, $4::text
);
