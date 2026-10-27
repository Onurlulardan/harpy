-- +goose Up
ALTER TABLE platform.billing_entitlement_registry_versions
    DROP CONSTRAINT billing_entitlement_registry_versions_shape_check;
ALTER TABLE platform.billing_entitlement_registry_versions
    ADD CONSTRAINT billing_entitlement_registry_versions_shape_check CHECK (
        registry_version > 0
        AND source_contract IN ('ADR-0082', 'ADR-0120')
        AND pg_catalog.octet_length(canonical_sha256) = 32
        AND definition_count > 0
    );

ALTER TABLE platform.billing_entitlement_definitions
    DROP CONSTRAINT billing_entitlement_definitions_key_check;
ALTER TABLE platform.billing_entitlement_definitions
    ADD CONSTRAINT billing_entitlement_definitions_key_check CHECK (
        entitlement_key ~ '^[a-z](?:[a-z0-9_.]{0,126}[a-z0-9])?$'
        AND owner_module ~ '^[a-z][a-z0-9_]{0,62}[a-z0-9]$'
        AND source_contract IN ('ADR-0082', 'ADR-0120')
        AND registry_version > 0
        AND pg_catalog.octet_length(canonical_sha256) = 32
    );

ALTER TABLE platform.billing_entitlement_registry_versions
    DISABLE TRIGGER billing_entitlement_registry_versions_authority_guard;
ALTER TABLE platform.billing_entitlement_definitions
    DISABLE TRIGGER billing_entitlement_definitions_authority_guard;

INSERT INTO platform.billing_entitlement_registry_versions (
    registry_version, authority_id, source_contract,
    canonical_sha256, definition_count, published_at
)
SELECT 2, authority_id, 'ADR-0120',
       pg_catalog.sha256(pg_catalog.convert_to(
           'analytics.active_dashboards|integer|gauge|never|hard|true|analytics|count
'
           || 'analytics.dashboard_builder.enabled|boolean||||false|analytics|enabled
'
           || 'api.active_endpoints|integer|gauge|never|hard|true|platform_api|count
'
           || 'api.builder.enabled|boolean||||false|platform_api|enabled
'
           || 'workflow.active_definitions|integer|gauge|never|hard|true|workflow|count
'
           || 'workflow.builder.enabled|boolean||||false|workflow|enabled',
           'UTF8'
       )), 6, '2026-07-22 00:00:00+00'::timestamptz
FROM platform.billing_global_authorities;

INSERT INTO platform.billing_entitlement_definitions (
    id, authority_id, registry_version, entitlement_key, value_type,
    meter_kind, reset_policy, enforcement, unlimited_allowed,
    owner_module, unit, source_contract, canonical_sha256, created_at
)
SELECT seed.id, registry.authority_id, 2, seed.entitlement_key,
       seed.value_type, seed.meter_kind, seed.reset_policy,
       seed.enforcement, seed.unlimited_allowed, seed.owner_module,
       seed.unit, 'ADR-0120',
       pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.concat_ws(
           '|', seed.entitlement_key, seed.value_type,
           COALESCE(seed.meter_kind, ''), COALESCE(seed.reset_policy, ''),
           COALESCE(seed.enforcement, ''), seed.unlimited_allowed::text,
           seed.owner_module, seed.unit, '2'
       ), 'UTF8')), registry.published_at
FROM platform.billing_entitlement_registry_versions AS registry
CROSS JOIN (VALUES
    ('01900000-0000-7000-8000-000000000009'::uuid,
     'workflow.builder.enabled', 'boolean', NULL::text, NULL::text,
     NULL::text, false, 'workflow', 'enabled'),
    ('01900000-0000-7000-8000-00000000000a'::uuid,
     'api.builder.enabled', 'boolean', NULL::text, NULL::text,
     NULL::text, false, 'platform_api', 'enabled'),
    ('01900000-0000-7000-8000-00000000000b'::uuid,
     'analytics.dashboard_builder.enabled', 'boolean', NULL::text,
     NULL::text, NULL::text, false, 'analytics', 'enabled'),
    ('01900000-0000-7000-8000-00000000000c'::uuid,
     'workflow.active_definitions', 'integer', 'gauge', 'never',
     'hard', true, 'workflow', 'count'),
    ('01900000-0000-7000-8000-00000000000d'::uuid,
     'api.active_endpoints', 'integer', 'gauge', 'never',
     'hard', true, 'platform_api', 'count'),
    ('01900000-0000-7000-8000-00000000000e'::uuid,
     'analytics.active_dashboards', 'integer', 'gauge', 'never',
     'hard', true, 'analytics', 'count')
) AS seed(
    id, entitlement_key, value_type, meter_kind, reset_policy,
    enforcement, unlimited_allowed, owner_module, unit
)
WHERE registry.registry_version = 2;

ALTER TABLE platform.billing_entitlement_registry_versions
    ENABLE ALWAYS TRIGGER billing_entitlement_registry_versions_authority_guard;
ALTER TABLE platform.billing_entitlement_definitions
    ENABLE ALWAYS TRIGGER billing_entitlement_definitions_authority_guard;
