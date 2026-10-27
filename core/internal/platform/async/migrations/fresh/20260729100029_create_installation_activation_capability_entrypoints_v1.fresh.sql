-- +goose Up
-- Migration stream version: 20260729100029.
CREATE TABLE platform.platform_async_installation_activation_receipts (
    operation_id uuid NOT NULL, environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL, job_id uuid, event_id uuid,
    receipt_kind text NOT NULL
        CHECK (receipt_kind IN ('initial_admin_setup', 'activation_event')),
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    PRIMARY KEY (operation_id, receipt_kind),
    UNIQUE (job_id), UNIQUE (event_id),
    CHECK (
        (receipt_kind = 'initial_admin_setup' AND job_id IS NOT NULL
            AND event_id IS NULL) OR
        (receipt_kind = 'activation_event' AND event_id IS NOT NULL
            AND job_id IS NULL)
    ));
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_artifact_id uuid, p_artifact_ref text, p_schema_name name,
    p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
SET application_name TO 'forgecrud-v2:core-operator-initialize'
AS $function$
DECLARE
    v_replayed boolean;
BEGIN
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id, 'activation.complete', p_environment_id,
        p_tenant_id, p_schema_name, 'saas_customer_tenant_provision'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'initial admin setup enqueue capability required';
    END IF;
    v_replayed := platform.platform_async_initial_admin_setup_enqueue_v1(
        p_operation_id, p_tenant_id, p_artifact_id, p_artifact_ref,
        p_schema_name, p_manifest_sha256
    );
    INSERT INTO platform.platform_async_installation_activation_receipts (
        operation_id, receipt_kind, environment_id, tenant_id, job_id,
        manifest_sha256
    ) VALUES (
        p_operation_id, 'initial_admin_setup', p_environment_id, p_tenant_id,
        p_artifact_id, p_manifest_sha256
    ) ON CONFLICT (operation_id, receipt_kind) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platform_async_installation_activation_receipts r
        WHERE r.operation_id = p_operation_id
          AND r.receipt_kind = 'initial_admin_setup'
          AND ROW(r.environment_id,r.tenant_id,r.job_id,r.manifest_sha256)
              = ROW(p_environment_id,p_tenant_id,p_artifact_id,p_manifest_sha256)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'initial admin setup activation receipt conflicts';
    END IF;
    RETURN v_replayed;
END;
$function$;
CREATE FUNCTION platform.platform_async_installation_activation_event_v2(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_actor_id uuid, p_schema_name name, p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
SET application_name TO 'forgecrud-v2:core-operator-initialize'
AS $function$
DECLARE
    v_replayed boolean;
    v_authorized boolean;
BEGIN
    v_authorized :=
        platform.tenancy_activation_step_context_assert_v1(
            p_operation_id, 'activation.complete', p_environment_id,
            p_tenant_id, p_schema_name, 'environment_initialize'
        ) OR platform.tenancy_activation_step_context_assert_v1(
            p_operation_id, 'activation.complete', p_environment_id,
            p_tenant_id, p_schema_name, 'saas_customer_tenant_provision'
        );
    IF NOT v_authorized THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation activation event capability required';
    END IF;
    v_replayed := platform.platform_async_installation_activation_event_v1(
        p_operation_id, p_environment_id, p_tenant_id, p_actor_id,
        p_schema_name, p_manifest_sha256
    );
    INSERT INTO platform.platform_async_installation_activation_receipts (
        operation_id, receipt_kind, environment_id, tenant_id, event_id,
        manifest_sha256
    ) VALUES (
        p_operation_id, 'activation_event', p_environment_id, p_tenant_id,
        p_operation_id, p_manifest_sha256
    ) ON CONFLICT (operation_id, receipt_kind) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platform_async_installation_activation_receipts r
        WHERE r.operation_id = p_operation_id
          AND r.receipt_kind = 'activation_event'
          AND ROW(r.environment_id,r.tenant_id,r.event_id,r.manifest_sha256)
              = ROW(p_environment_id,p_tenant_id,p_operation_id,p_manifest_sha256)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'installation activation event receipt conflicts';
    END IF;
    RETURN v_replayed;
END;
$function$;
CREATE FUNCTION platform.platform_async_installation_activation_evidence_assert_v1(
    p_operation_id uuid, p_expected_artifact_id uuid)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN SESSION_USER = 'forgecrud_v2_runtime'
       AND CURRENT_USER = 'forgecrud_owner'
       AND EXISTS (
        SELECT 1
        FROM platform.platform_async_installation_activation_receipts event
        JOIN platform.outbox outbox ON outbox.event_id = event.event_id
        LEFT JOIN platform.platform_async_installation_activation_receipts setup
          ON setup.operation_id = event.operation_id
         AND setup.receipt_kind = 'initial_admin_setup'
         AND ROW(setup.environment_id,setup.tenant_id,setup.manifest_sha256)
             = ROW(event.environment_id,event.tenant_id,event.manifest_sha256)
        LEFT JOIN platform.jobs job ON job.id = setup.job_id
        WHERE event.operation_id = p_operation_id
          AND event.receipt_kind = 'activation_event'
          AND ROW(outbox.tenant_id,outbox.scope_kind,outbox.scope_id,
              outbox.writer_origin,outbox.event_type,outbox.aggregate_type,
              outbox.aggregate_id,outbox.causation_id,outbox.idempotency_key,
              outbox.business_effect_idempotency_key)
          = ROW(event.tenant_id,'tenant',event.tenant_id,'v2',
              'installation.activated','installation.environment',
              event.tenant_id,event.operation_id,
              'installation.activation:' || event.operation_id::text,
              'installation.activation:' || event.operation_id::text)
          AND outbox.payload_canonical = pg_catalog.convert_to(
              pg_catalog.jsonb_build_object(
                'environment_id',event.environment_id::text,
                'manifest_sha256',event.manifest_sha256,
                'operation_id',event.operation_id::text,
                'tenant_id',event.tenant_id::text
              )::text,'UTF8'
          )
          AND (p_expected_artifact_id IS NULL OR (
              setup.job_id = p_expected_artifact_id
          AND ROW(job.tenant_id,job.scope_kind,job.scope_id,
              job.authority_group_id,job.authority_group_generation,
              job.authority_domain,job.cutover_generation,job.writer_origin,
              job.kind,job.payload_ref,job.dedupe_key,job.effect_class,
              job.effect_idempotency_key)
          = ROW(setup.tenant_id,'tenant',setup.tenant_id,
              'iam_system_resources',1::bigint,'iam',1::bigint,'v2',
              'iam.lifecycle.delivery',
              'iam-lifecycle-delivery://tenant/' || setup.tenant_id::text ||
                  '/artifact/' || setup.job_id::text,
              setup.job_id::text,'external',
              'iam.lifecycle.delivery:iam-lifecycle-delivery://tenant/' ||
                  setup.tenant_id::text || '/artifact/' || setup.job_id::text)
          ))
       );
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(
    uuid,uuid,uuid,uuid,text,name,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(
    uuid,uuid,uuid,uuid,text,name,text) TO forgecrud_runtime;
REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v2(
    uuid,uuid,uuid,uuid,name,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v2(
    uuid,uuid,uuid,uuid,name,text) TO forgecrud_runtime;
REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_evidence_assert_v1(
    uuid,uuid) FROM PUBLIC;
