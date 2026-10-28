-- +goose Up
-- Migration stream version: 20260728110020.

WITH required_runtime(signature) AS (
    VALUES
        ('platform.billing_resolve_current_tenant_authority_v1(uuid)'),
        ('platform.billing_resolve_current_global_authority_v1(text)'),
        ('platform.billing_effective_subscription_price_snapshot_read_v1(uuid,uuid,uuid,timestamptz)'),
        ('platform.billing_subscription_snapshot_read_v1(uuid,uuid)'),
        ('platform.billing_usage_context_snapshot_read_v1(uuid,uuid)'),
        ('platform.billing_usage_snapshot_read_v1(uuid,uuid,uuid,uuid,uuid)'),
        ('platform.billing_discount_eligibility_snapshot_read_v1(uuid,uuid,text,uuid,timestamptz,uuid)'),
        ('platform.billing_discount_offerings_snapshot_read_v1(uuid,uuid)')
), owner_only(signature) AS (
    VALUES
        ('platform.analytics_resolve_current_tenant_authority_v1(uuid)')
), lock_free(signature) AS (
    SELECT signature FROM required_runtime
    UNION ALL
    SELECT signature FROM owner_only
), resolved AS (
    SELECT signature,
           pg_catalog.to_regprocedure(signature) AS procedure_oid
    FROM lock_free
), required_views(relation_name) AS (
    VALUES
        ('platform.billing_catalog_offering_snapshot_read_v1'),
        ('platform.billing_catalog_campaign_snapshot_read_v1')
), owner_tables(relation_name) AS (
    VALUES
        ('platform.billing_global_authorities'),
        ('platform.billing_tenant_authorities'),
        ('platform.billing_offerings'),
        ('platform.billing_offering_versions'),
        ('platform.billing_price_versions'),
        ('platform.billing_subscriptions'),
        ('platform.billing_usage_totals'),
        ('platform.billing_discounts')
)
SELECT 1 / CASE WHEN
    EXISTS (
        SELECT 1
        FROM resolved
        LEFT JOIN pg_catalog.pg_proc AS procedure_row
          ON procedure_row.oid = resolved.procedure_oid
        WHERE resolved.procedure_oid IS NULL
           OR procedure_row.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
           OR NOT procedure_row.prosecdef
           OR NOT (
               COALESCE(procedure_row.proconfig, ARRAY[]::text[])
               @> ARRAY['search_path=pg_catalog']::text[]
           )
           OR pg_catalog.pg_get_functiondef(resolved.procedure_oid)
              ~* 'FOR[[:space:]]+(SHARE|UPDATE)'
    )
    OR EXISTS (
        SELECT 1
        FROM required_runtime
        WHERE NOT pg_catalog.has_function_privilege(
            'harpy_runtime',
            pg_catalog.to_regprocedure(required_runtime.signature),
            'EXECUTE'
        )
    )
    OR EXISTS (
        SELECT 1
        FROM owner_only
        WHERE pg_catalog.has_function_privilege(
            'harpy_runtime',
            pg_catalog.to_regprocedure(owner_only.signature),
            'EXECUTE'
        )
    )
    OR EXISTS (
        SELECT 1
        FROM required_views
        LEFT JOIN pg_catalog.pg_class AS relation
          ON relation.oid = pg_catalog.to_regclass(required_views.relation_name)
        WHERE relation.oid IS NULL
           OR relation.relkind IS DISTINCT FROM 'v'
           OR NOT (
               COALESCE(relation.reloptions, ARRAY[]::text[])
               @> ARRAY['security_barrier=true']::text[]
           )
           OR NOT pg_catalog.has_table_privilege(
               'harpy_runtime', relation.oid, 'SELECT'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM owner_tables
        WHERE pg_catalog.to_regclass(owner_tables.relation_name) IS NULL
           OR pg_catalog.has_table_privilege(
               'harpy_runtime',
               pg_catalog.to_regclass(owner_tables.relation_name),
               'SELECT'
           )
    )
THEN 0 ELSE 1 END;
