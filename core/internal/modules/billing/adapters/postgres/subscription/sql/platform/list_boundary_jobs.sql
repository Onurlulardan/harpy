SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
    'id', intent.id,
    'effect_id', intent.effect_id,
    'subscription_id', intent.subscription_id,
    'boundary', intent.boundary_kind,
    'run_at_unix_micro',
    (EXTRACT(EPOCH FROM intent.run_at) * 1000000)::bigint,
    'expected_revision', intent.expected_revision,
    'payload_revision', intent.payload_revision
) ORDER BY intent.run_at, intent.id), '[]'::jsonb)
FROM platform.billing_subscription_boundary_jobs AS intent
JOIN platform.billing_tenant_authorities AS authority
  ON authority.authority_id = intent.authority_id
 AND authority.tenant_id = intent.tenant_id
JOIN platform.jobs AS job
  ON job.id = intent.id AND job.scope_kind = 'tenant'
 AND job.scope_id = intent.tenant_id AND job.tenant_id = intent.tenant_id
 AND job.authority_group_id = authority.authority_group_id
 AND job.authority_group_generation = authority.authority_group_generation
 AND job.authority_domain = authority.billing_domain
 AND job.cutover_generation = authority.billing_generation
 AND job.payload_revision = intent.payload_revision
 AND job.payload_ref = intent.payload_reference
 AND job.payload_sha256 = pg_catalog.encode(intent.payload_sha256, 'hex')
WHERE intent.authority_id = $1 AND intent.tenant_id = $2
  AND intent.subscription_id = $3 AND job.state IN ('queued', 'running');
