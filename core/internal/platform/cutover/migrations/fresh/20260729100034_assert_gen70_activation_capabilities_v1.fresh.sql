-- +goose Up
-- Migration stream version: 20260729100034.

WITH public_routines(signature) AS (
    VALUES
        ('platform.tenancy_onboarding_activate_v3(uuid,text,bigint)'),
        ('platform.tenancy_onboarding_activate_v2(uuid,text,bigint)'),
        ('platform.tenancy_activation_initialize_v5(uuid,uuid,uuid,text,text,uuid,text,text,text,text)'),
        ('platform.tenancy_authority_reconcile_current_v1(uuid,text)'),
        ('platform.billing_installation_assign_v3(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)'),
        ('platform.iam_installation_activation_assert_v2(uuid,uuid,uuid,name)'),
        ('platform.iam_installation_security_epoch_initialize_v2(uuid,uuid,uuid,name,timestamptz)'),
        ('platform.iam_installation_identity_directory_upsert_v1(uuid,uuid,uuid,name,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea)'),
        ('platform.iam_installation_instance_owner_bind_v2(uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text)'),
        ('platform.billing_installation_open_usage_v2(uuid,uuid,text)'),
        ('platform.platformops_installation_recovery_initialize_v2(uuid,uuid,uuid,name,text,bigint,text,text)'),
        ('platform.platform_async_initial_admin_setup_enqueue_v2(uuid,uuid,uuid,uuid,text,name,text)'),
        ('platform.platform_async_installation_activation_event_v2(uuid,uuid,uuid,uuid,name,text)'),
        ('platform.platformauthority_installation_owner_bind_v3(uuid,uuid,uuid,name,text,uuid,uuid,uuid,text)'),
        ('platform.tenancy_onboarding_complete_v2(uuid,uuid,text,text,bigint)'),
        ('platform.platformauthority_installation_head_pin_v3(uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text)'),
        ('platform.tenancy_onboarding_accept_v2(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)')
), internal_routines(signature) AS (
    VALUES
        ('platform.tenancy_activation_step_assert_v1(uuid,text)'),
        ('platform.tenancy_activation_step_consume_v1(uuid,text)'),
        ('platform.tenancy_activation_step_context_assert_v1(uuid,text,uuid,uuid,name,text)'),
        ('platform.tenancy_activation_step_commit_guard_v1()'),
        ('platform.platform_async_installation_activation_evidence_assert_v1(uuid,uuid)'),
        ('platform.tenancy_activation_command_consume_v1(uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text,timestamptz,bigint,text)'),
        ('platform.tenancy_authority_reconcile_scope_assert_v1(text)'),
        ('platform.tenancy_activation_capability_commit_guard_v1()'),
        ('platform.tenancy_authority_scope_commit_guard_v1()'),
        ('platform.billing_installation_id_v1(uuid,text)'),
        ('platform.analytics_reconcile_current_tenant_authorities_v2()'),
        ('platform.billing_reconcile_current_global_authority_v2()'),
        ('platform.billing_reconcile_current_tenant_authorities_v2()')
), legacy_routines(signature) AS (
    VALUES
        ('platform.tenancy_onboarding_activate_v1(uuid,text,bigint)'),
        ('platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text)'),
        ('platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)'),
        ('platform.billing_reconcile_current_global_authority_v1()'),
        ('platform.billing_reconcile_current_tenant_authorities_v1()'),
        ('platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz)'),
        ('platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,bigint,timestamptz)'),
        ('platform.platformauthority_installation_head_pin_v2(uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text)'),
        ('platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text)'),
        ('platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint)'),
        ('platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)'),
        ('platform.billing_installation_open_usage_v1(uuid,uuid,text)'),
        ('platform.platformops_installation_recovery_initialize_v1(uuid,uuid,uuid,name,text,bigint,text,text)'),
        ('platform.platform_async_initial_admin_setup_enqueue_v1(uuid,uuid,uuid,text,name,text)'),
        ('platform.platform_async_installation_activation_event_v1(uuid,uuid,uuid,uuid,name,text)'),
        ('platform.platformauthority_installation_owner_bind_v2(uuid,uuid,uuid,name,text,uuid,uuid,uuid,text)'),
        ('platform.tenancy_activation_assert_v1(uuid,uuid,name)'),
        ('platform.tenancy_activation_assert_v2(uuid,uuid,name)'),
        ('platform.iam_security_epoch_initialize_v1(uuid,timestamptz)'),
        ('platform.iam_installation_instance_owner_bind_v1(uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text)'),
        ('platform.iam_installation_instance_owner_input_v1(uuid)')
), public_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM public_routines
), internal_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM internal_routines
), legacy_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM legacy_routines
), required_tables(relation_name) AS (
    VALUES
        ('platform.tenancy_activation_step_capabilities'),
        ('platform.platform_async_installation_activation_receipts'),
        ('platform.tenancy_activation_command_capabilities'),
        ('platform.tenancy_authority_reconcile_scopes'),
        ('platform.billing_installation_assignment_contexts')
), table_resolved AS (
    SELECT relation_name,
           pg_catalog.to_regclass(relation_name) AS relation_oid
    FROM required_tables
), forbidden_roles(role_name) AS (
    VALUES ('harpy_runtime'), ('harpy_v2_runtime')
), forbidden_privileges(privilege_name) AS (
    VALUES
        ('SELECT'), ('INSERT'), ('UPDATE'), ('DELETE'),
        ('TRUNCATE'), ('REFERENCES'), ('TRIGGER')
)
SELECT 1 / CASE WHEN
    EXISTS (
        SELECT 1
        FROM public_resolved
        LEFT JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = public_resolved.routine_oid
        WHERE routine_oid IS NULL
           OR routine.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
           OR NOT routine.prosecdef
           OR NOT (
               COALESCE(routine.proconfig, ARRAY[]::text[])
               @> ARRAY['search_path=pg_catalog']::text[]
           )
           OR NOT pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM internal_resolved
        LEFT JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = internal_resolved.routine_oid
        WHERE routine_oid IS NULL
           OR routine.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
           OR NOT routine.prosecdef
           OR NOT (
               COALESCE(routine.proconfig, ARRAY[]::text[])
               @> ARRAY['search_path=pg_catalog']::text[]
           )
           OR pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM legacy_resolved
        WHERE routine_oid IS NULL
           OR NOT pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM table_resolved
        LEFT JOIN pg_catalog.pg_class AS relation
          ON relation.oid = table_resolved.relation_oid
        WHERE relation_oid IS NULL
           OR relation.relkind IS DISTINCT FROM 'r'
           OR relation.relowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
    )
    OR EXISTS (
        SELECT 1
        FROM table_resolved
        CROSS JOIN forbidden_roles
        CROSS JOIN forbidden_privileges
        WHERE pg_catalog.has_table_privilege(
            role_name, relation_oid, privilege_name
        )
    )
THEN 0 ELSE 1 END;
