-- +goose Up
INSERT INTO platform.billing_entitlement_registry_versions (
    registry_version, authority_id, source_contract,
    canonical_sha256, definition_count, published_at
)
SELECT 1, authority_id, 'ADR-0082',
       pg_catalog.sha256(pg_catalog.convert_to(
           'api.access|boolean||||false|platform_api|enabled\n'
           || 'documents.storage_bytes|integer|gauge|never|hard|true|documents|byte\n'
           || 'forms.active_forms|integer|gauge|never|hard|true|forms|count\n'
           || 'forms.submissions|integer|counter|calendar_month_utc|hard|true|forms|count\n'
           || 'iam.active_users|integer|gauge|never|hard|true|iam|count\n'
           || 'workflow.enabled|boolean||||false|workflow|enabled',
           'UTF8'
       )),
       6, '2026-07-17 00:00:00+00'::timestamptz
FROM platform.billing_global_authorities;

INSERT INTO platform.billing_entitlement_definitions (
    id, authority_id, registry_version, entitlement_key, value_type,
    meter_kind, reset_policy, enforcement, unlimited_allowed,
    owner_module, unit, source_contract, canonical_sha256, created_at
)
SELECT seed.id, registry.authority_id, registry.registry_version,
       seed.entitlement_key, seed.value_type, seed.meter_kind,
       seed.reset_policy, seed.enforcement, seed.unlimited_allowed,
       seed.owner_module, seed.unit, registry.source_contract,
       pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.concat_ws(
           '|', seed.entitlement_key, seed.value_type,
           COALESCE(seed.meter_kind, ''),
           COALESCE(seed.reset_policy, ''),
           COALESCE(seed.enforcement, ''),
           seed.unlimited_allowed::text, seed.owner_module, seed.unit,
           registry.registry_version::text
       ), 'UTF8')),
       registry.published_at
FROM platform.billing_entitlement_registry_versions AS registry
CROSS JOIN (VALUES
    ('01900000-0000-7000-8000-000000000002'::uuid,
     'api.access', 'boolean', NULL::text, NULL::text, NULL::text,
     false, 'platform_api', 'enabled'),
    ('01900000-0000-7000-8000-000000000003'::uuid,
     'documents.storage_bytes', 'integer', 'gauge', 'never', 'hard',
     true, 'documents', 'byte'),
    ('01900000-0000-7000-8000-000000000004'::uuid,
     'forms.active_forms', 'integer', 'gauge', 'never', 'hard',
     true, 'forms', 'count'),
    ('01900000-0000-7000-8000-000000000005'::uuid,
     'forms.submissions', 'integer', 'counter', 'calendar_month_utc', 'hard',
     true, 'forms', 'count'),
    ('01900000-0000-7000-8000-000000000006'::uuid,
     'iam.active_users', 'integer', 'gauge', 'never', 'hard',
     true, 'iam', 'count'),
    ('01900000-0000-7000-8000-000000000007'::uuid,
     'workflow.enabled', 'boolean', NULL::text, NULL::text, NULL::text,
     false, 'workflow', 'enabled')
) AS seed(
    id, entitlement_key, value_type, meter_kind, reset_policy,
    enforcement, unlimited_allowed, owner_module, unit
)
WHERE registry.registry_version = 1;
