-- +goose Up
-- +goose StatementBegin
REVOKE ALL ON FUNCTION platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text) FROM PUBLIC;
CREATE FUNCTION platform.billing_installation_id_v1(
    p_operation_id uuid, p_label text)
RETURNS uuid LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_bytes bytea;
BEGIN
    IF p_operation_id IS NULL OR p_label IS NULL
       OR p_label NOT IN ('tenant-authority','base-offering-assignment') THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing installation identifier label is invalid';
    END IF;
    v_bytes := pg_catalog.substring(pg_catalog.sha256(
        pg_catalog.convert_to('harpy:billing:installation:v1','UTF8')
        || pg_catalog.uuid_send(p_operation_id)
        || pg_catalog.decode('00','hex')
        || pg_catalog.convert_to(p_label,'UTF8')),1,16);
    v_bytes := pg_catalog.set_byte(v_bytes, 6,
        (pg_catalog.get_byte(v_bytes, 6) & 15) | 112);
    v_bytes := pg_catalog.set_byte(v_bytes, 8,
        (pg_catalog.get_byte(v_bytes, 8) & 63) | 128);
    RETURN pg_catalog.encode(v_bytes,'hex')::uuid;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_installation_id_v1(uuid,text) FROM PUBLIC;
CREATE TABLE platform.billing_installation_assignment_contexts (
    operation_id uuid PRIMARY KEY REFERENCES platform.billing_installation_base_assignments(operation_id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    deployment_mode text NOT NULL,
    operation_kind text NOT NULL,
    binding_purpose text NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT billing_installation_assignment_contexts_profile_ck CHECK (
        (operation_kind='environment_initialize' AND deployment_mode='onpremise'
         AND binding_purpose='onpremise_primary')
        OR (operation_kind='environment_initialize'
            AND deployment_mode='saas' AND binding_purpose='saas_control')
        OR (operation_kind='saas_customer_tenant_provision'
            AND deployment_mode='saas' AND binding_purpose='saas_customer')
    ));
CREATE TRIGGER billing_installation_assignment_contexts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.billing_installation_assignment_contexts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
INSERT INTO platform.billing_installation_assignment_contexts (
    operation_id,deployment_mode,operation_kind,binding_purpose,recorded_at
)
SELECT fact.operation_id,fact.deployment_mode,
        CASE WHEN fact.deployment_mode='onpremise' OR control.match_count=1
             THEN 'environment_initialize'
             ELSE 'saas_customer_tenant_provision' END,
        CASE WHEN fact.deployment_mode='onpremise' THEN 'onpremise_primary'
             WHEN control.match_count=1 THEN 'saas_control'
             ELSE 'saas_customer' END,
        fact.assigned_at
    FROM platform.billing_installation_base_assignments AS fact
    CROSS JOIN LATERAL (
        SELECT pg_catalog.count(*) AS match_count
        FROM platform.billing_license_offering_publications AS publication
        WHERE ROW(publication.environment_id,publication.offering_version_id,
            publication.revision,publication.offering_sha256,
            publication.effective_at,publication.snapshot_sequence,
            publication.snapshot_sha256) IS NOT DISTINCT FROM ROW(
            fact.environment_id,fact.offering_version_id,fact.offering_revision,
            fact.offering_sha256,fact.offering_effective_at,
            fact.source_sequence,fact.source_sha256)
    ) AS license
    CROSS JOIN LATERAL (
        SELECT pg_catalog.count(*) AS match_count
        FROM platform.billing_saas_control_offering_publications AS publication
        WHERE publication.offering_key='saas_platform_control_base_v1'
          AND ROW(publication.offering_version_id,publication.revision,
            publication.offering_sha256,publication.effective_at)
            IS NOT DISTINCT FROM ROW(fact.offering_version_id,
            fact.offering_revision,fact.offering_sha256,
            fact.offering_effective_at)
    ) AS control
    CROSS JOIN LATERAL (
        SELECT pg_catalog.count(*) AS match_count
        FROM platform.billing_saas_free_base_publications AS publication
        WHERE publication.offering_key='saas_free_base_v1'
          AND ROW(publication.offering_version_id,publication.revision,
            publication.offering_sha256,publication.effective_at)
            IS NOT DISTINCT FROM ROW(fact.offering_version_id,
            fact.offering_revision,fact.offering_sha256,
            fact.offering_effective_at)
    ) AS customer
    WHERE (
        fact.deployment_mode='onpremise' AND license.match_count=1
        OR fact.deployment_mode='saas'
           AND control.match_count+customer.match_count=1
    )
      AND fact.authority_id=platform.billing_installation_id_v1(
        fact.operation_id,'tenant-authority')
  AND fact.assignment_id=platform.billing_installation_id_v1(
      fact.operation_id,'base-offering-assignment');
CREATE FUNCTION platform.billing_installation_assign_v3(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_authority_id uuid, p_assignment_id uuid, p_deployment_mode text,
    p_offering_version_id uuid, p_offering_revision bigint,
    p_offering_sha256 text, p_offering_effective_at timestamptz,
    p_source_sequence bigint, p_source_sha256 text, p_manifest_sha256 text,
    p_binding_purpose text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_context platform.billing_installation_assignment_contexts%ROWTYPE;
    v_operation_kind text;
    v_context_found boolean;
    v_replayed boolean;
BEGIN
    IF p_deployment_mode='onpremise'
       AND p_binding_purpose='onpremise_primary' THEN
        v_operation_kind := 'environment_initialize';
    ELSIF p_deployment_mode='saas'
          AND p_binding_purpose='saas_control' THEN
        v_operation_kind := 'environment_initialize';
    ELSIF p_deployment_mode='saas'
          AND p_binding_purpose='saas_customer' THEN
        v_operation_kind := 'saas_customer_tenant_provision';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing installation assignment profile is invalid';
    END IF;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_authority_id IS DISTINCT FROM
            platform.billing_installation_id_v1(
                p_operation_id,'tenant-authority')
       OR p_assignment_id IS DISTINCT FROM
            platform.billing_installation_id_v1(
                p_operation_id,'base-offering-assignment')
       OR platform.tenancy_activation_command_consume_v1(
           p_operation_id,'billing.assign_tenant_base_offering',
           p_environment_id,p_tenant_id,v_operation_kind,p_deployment_mode,
           p_binding_purpose,p_manifest_sha256,p_offering_version_id,
           p_offering_revision,p_offering_sha256,p_offering_effective_at,
           p_source_sequence,p_source_sha256
       ) IS DISTINCT FROM TRUE THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing installation assignment capability is denied';
    END IF;
    SELECT * INTO v_context
    FROM platform.billing_installation_assignment_contexts AS context
    WHERE context.operation_id=p_operation_id FOR SHARE;
    v_context_found := FOUND;
    IF v_context_found AND ROW(
        v_context.deployment_mode,v_context.operation_kind,
        v_context.binding_purpose
    ) IS DISTINCT FROM ROW(
        p_deployment_mode,v_operation_kind,p_binding_purpose
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'billing installation assignment context conflicts';
    END IF;
    v_replayed := platform.billing_installation_assign_v2(
        p_operation_id,p_environment_id,p_tenant_id,p_authority_id,
        p_assignment_id,p_deployment_mode,p_offering_version_id,
        p_offering_revision,p_offering_sha256,p_offering_effective_at,
        p_source_sequence,p_source_sha256,p_manifest_sha256,p_binding_purpose
    );
    IF v_replayed IS DISTINCT FROM v_context_found THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing installation assignment replay differs';
    END IF;
    IF NOT v_context_found THEN
        INSERT INTO platform.billing_installation_assignment_contexts (
            operation_id,deployment_mode,operation_kind,binding_purpose,recorded_at
        ) VALUES (
            p_operation_id,p_deployment_mode,v_operation_kind,
            p_binding_purpose,pg_catalog.transaction_timestamp()
        );
    END IF;
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.billing_installation_assign_v3(uuid,uuid,uuid,
    uuid,uuid,text,uuid,bigint,text,timestamptz,
    bigint,text,text,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_assign_v3(uuid,uuid,
    uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text) TO harpy_runtime;
