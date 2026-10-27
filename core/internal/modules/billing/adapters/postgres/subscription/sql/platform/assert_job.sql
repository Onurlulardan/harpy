SELECT intent.effect_id, intent.subscription_id, intent.boundary_kind,
       intent.run_at, intent.expected_revision, intent.payload_revision
FROM platform.billing_subscription_boundary_jobs AS intent
JOIN platform.billing_tenant_authorities AS authority
  ON authority.authority_id = intent.authority_id
 AND authority.tenant_id = intent.tenant_id
JOIN platform.jobs AS job ON job.id = intent.id
WHERE intent.authority_id = $1 AND intent.tenant_id = $2 AND intent.id = $3
  AND intent.subscription_id = $4 AND intent.effect_id = $5
  AND intent.expected_revision = $6 AND intent.boundary_kind = $7
  AND job.scope_kind = 'tenant' AND job.scope_id = intent.tenant_id
  AND job.tenant_id = intent.tenant_id AND job.state = 'running'
  AND job.lease_generation = $8
  AND job.authority_group_id = authority.authority_group_id
  AND job.authority_group_generation = authority.authority_group_generation
  AND job.authority_domain = authority.billing_domain
  AND job.cutover_generation = authority.billing_generation
  AND job.payload_revision = intent.payload_revision
  AND job.payload_ref = intent.payload_reference
  AND job.payload_sha256 = pg_catalog.encode(intent.payload_sha256, 'hex')
FOR SHARE OF intent, job;
