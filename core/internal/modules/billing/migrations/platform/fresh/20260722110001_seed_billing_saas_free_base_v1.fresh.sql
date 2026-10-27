-- +goose Up
ALTER TABLE platform.billing_offerings DISABLE TRIGGER billing_offerings_authority_guard;
ALTER TABLE platform.billing_offering_versions DISABLE TRIGGER billing_offering_versions_authority_guard;
ALTER TABLE platform.billing_price_versions DISABLE TRIGGER billing_price_versions_authority_guard;
ALTER TABLE platform.billing_offering_entitlements DISABLE TRIGGER billing_offering_entitlements_authority_guard;

INSERT INTO platform.billing_offerings (
    id,authority_id,offering_key,lifecycle,created_at,retired_at,revision
) SELECT '019b0000-0000-7000-8000-000000000200',authority_id,
    'saas_free_base_v1','active','2026-07-22T00:00:00Z',NULL,1
FROM platform.billing_global_authorities;

INSERT INTO platform.billing_offering_versions (
    id,authority_id,offering_id,version,is_base,is_public,available,
    trial_days,grace_days,expected_price_count,expected_grant_count,
    effective_start,effective_end,presentation_metadata,presentation_sha256,
    publication_state,publication_operation_id,published_at
) SELECT '019b0000-0000-7000-8000-000000000201',authority_id,
    '019b0000-0000-7000-8000-000000000200',1,true,false,true,0,0,1,14,
    '2026-07-22T00:00:00Z',NULL,metadata,
    pg_catalog.sha256(pg_catalog.convert_to(metadata::text,'UTF8')),
    'published','019b0000-0000-7000-8000-000000000202','2026-07-22T00:00:00Z'
FROM platform.billing_global_authorities
CROSS JOIN LATERAL (SELECT pg_catalog.jsonb_build_object(
    'name','ForgeCRUD Free','description','Release-owned SaaS customer base',
    'sort_order',1,'highlight_label','','color','','billing_interval','monthly'
) AS metadata) AS presentation;

INSERT INTO platform.billing_price_versions (
    id,authority_id,offering_version_id,cadence,minor_units,currency,
    currency_registry_version,currency_eligible,effective_start,effective_end,
    money_policy_version,published_at
) SELECT '019b0000-0000-7000-8000-000000000203',authority_id,
    '019b0000-0000-7000-8000-000000000201','monthly',0,'USD',1,true,
    '2026-07-22T00:00:00Z',NULL,'fc_money_v1','2026-07-22T00:00:00Z'
FROM platform.billing_global_authorities;

INSERT INTO platform.billing_offering_entitlements (
    authority_id,offering_version_id,entitlement_definition_id,value_type,
    definition_allows_unlimited,boolean_value,integer_value,unlimited,created_at
) SELECT authority.authority_id,'019b0000-0000-7000-8000-000000000201',
    definition.id,definition.value_type,definition.unlimited_allowed,
    CASE definition.entitlement_key
      WHEN 'ai.chat' THEN false WHEN 'workflow.enabled' THEN false
      WHEN 'api.access' THEN false WHEN 'forms.builder.enabled' THEN false
      WHEN 'workflow.builder.enabled' THEN false WHEN 'api.builder.enabled' THEN false
      WHEN 'analytics.dashboard_builder.enabled' THEN false ELSE NULL END,
    CASE definition.entitlement_key
      WHEN 'forms.submissions' THEN 500 WHEN 'iam.active_users' THEN 3
      WHEN 'documents.storage_bytes' THEN 104857600
      WHEN 'workflow.active_definitions' THEN 0 WHEN 'api.active_endpoints' THEN 0
      WHEN 'analytics.active_dashboards' THEN 0 ELSE NULL END,
    definition.entitlement_key='forms.active_forms','2026-07-22T00:00:00Z'
FROM platform.billing_global_authorities AS authority
CROSS JOIN platform.billing_entitlement_definitions AS definition;

SELECT 1 / ((pg_catalog.count(*)=14)::integer)
FROM platform.billing_offering_entitlements
WHERE offering_version_id='019b0000-0000-7000-8000-000000000201';

INSERT INTO platform.billing_saas_free_base_publications (
    offering_key,offering_version_id,revision,offering_sha256,
    offering_document,effective_at,created_at
) SELECT 'saas_free_base_v1','019b0000-0000-7000-8000-000000000201',1,
    pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(document::text,'UTF8')),'hex'),
    document,'2026-07-22T00:00:00Z',pg_catalog.statement_timestamp()
FROM (SELECT pg_catalog.jsonb_build_object(
    'schema','forgecrud-saas-free-base.v1','offering_key','saas_free_base_v1',
    'revision',1,'effective_at','2026-07-22T00:00:00Z','subscription',false,'trial',false,
    'entitlements',pg_catalog.jsonb_build_object(
        'forms.submissions',500,'iam.active_users',3,'forms.active_forms','unlimited',
        'documents.storage_bytes',104857600,'ai.chat',false,'workflow.enabled',false,
        'api.access',false,'forms.builder.enabled',false,'workflow.builder.enabled',false,
        'api.builder.enabled',false,'analytics.dashboard_builder.enabled',false,
        'workflow.active_definitions',0,'api.active_endpoints',0,
        'analytics.active_dashboards',0)) AS document) AS publication;
