-- +goose Up
ALTER TABLE platform.billing_offerings
    DISABLE TRIGGER billing_offerings_authority_guard;
ALTER TABLE platform.billing_offering_versions
    DISABLE TRIGGER billing_offering_versions_authority_guard;
ALTER TABLE platform.billing_price_versions
    DISABLE TRIGGER billing_price_versions_authority_guard;
ALTER TABLE platform.billing_offering_entitlements
    DISABLE TRIGGER billing_offering_entitlements_authority_guard;

INSERT INTO platform.billing_offerings (
    id,authority_id,offering_key,lifecycle,created_at,retired_at,revision
) VALUES (
    '019b0000-0000-7000-8000-000000000100',
    (SELECT authority_id FROM platform.billing_global_authorities),
    'saas_platform_control_base_v1','active',
    '2026-07-22T00:00:00Z'::timestamptz,NULL,1
);

INSERT INTO platform.billing_offering_versions (
    id,authority_id,offering_id,version,is_base,is_public,available,
    trial_days,grace_days,expected_price_count,expected_grant_count,
    effective_start,effective_end,presentation_metadata,presentation_sha256,
    publication_state,publication_operation_id,published_at
) SELECT
    '019b0000-0000-7000-8000-000000000101',authority_id,
    '019b0000-0000-7000-8000-000000000100',1,true,false,false,
    0,0,1,13,effective_at,NULL,metadata,
    pg_catalog.sha256(pg_catalog.convert_to(metadata::text,'UTF8')),
    'published','019b0000-0000-7000-8000-000000000102',effective_at
FROM (
    SELECT authority_id,'2026-07-22T00:00:00Z'::timestamptz AS effective_at,
        pg_catalog.jsonb_build_object(
            'name','Harpy SaaS Control',
            'description','Platform control tenant base','sort_order',0,
            'highlight_label','','color','','billing_interval','monthly'
        ) AS metadata
    FROM platform.billing_global_authorities
) AS seed;

INSERT INTO platform.billing_price_versions (
    id,authority_id,offering_version_id,cadence,minor_units,currency,
    currency_registry_version,currency_eligible,effective_start,effective_end,
    money_policy_version,published_at
) SELECT
    '019b0000-0000-7000-8000-000000000103',authority_id,
    '019b0000-0000-7000-8000-000000000101','one_time',0,'USD',1,true,
    '2026-07-22T00:00:00Z'::timestamptz,NULL,'fc_money_v1',
    '2026-07-22T00:00:00Z'::timestamptz
FROM platform.billing_global_authorities;

INSERT INTO platform.billing_offering_entitlements (
    authority_id,offering_version_id,entitlement_definition_id,value_type,
    definition_allows_unlimited,boolean_value,integer_value,unlimited,created_at
) SELECT authority.authority_id,'019b0000-0000-7000-8000-000000000101',
    definition.id,definition.value_type,definition.unlimited_allowed,
    CASE WHEN definition.value_type='boolean'
        THEN definition.entitlement_key='api.access' ELSE NULL END,
    CASE WHEN definition.value_type='integer'
              AND definition.entitlement_key<>'iam.active_users' THEN 0 ELSE NULL END,
    definition.entitlement_key='iam.active_users',
    '2026-07-22T00:00:00Z'::timestamptz
FROM platform.billing_entitlement_definitions AS definition
CROSS JOIN platform.billing_global_authorities AS authority;

INSERT INTO platform.billing_saas_control_offering_publications (
    offering_key,offering_version_id,revision,offering_sha256,
    offering_document,effective_at,created_at
) SELECT 'saas_platform_control_base_v1',
    '019b0000-0000-7000-8000-000000000101',1,
    pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(document::text,'UTF8')),'hex'),
    document,'2026-07-22T00:00:00Z'::timestamptz,
    pg_catalog.statement_timestamp()
FROM (
    SELECT pg_catalog.jsonb_build_object(
        'schema','harpy-saas-control-offering.v1',
        'offering_key','saas_platform_control_base_v1','revision',1,
        'effective_at','2026-07-22T00:00:00Z',
        'policy','control_tenant_product_admission_denied',
        'entitlements',pg_catalog.jsonb_build_object(
            'iam.active_users','unlimited','api.access',true,
            'documents.storage_bytes',0,
            'workflow.enabled',false,
            'workflow.builder.enabled',false,'api.builder.enabled',false,
            'analytics.dashboard_builder.enabled',false,
            'workflow.active_definitions',0,'api.active_endpoints',0,
            'analytics.active_dashboards',0
        )
    ) AS document
) AS seed;
