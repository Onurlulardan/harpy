-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_assign_v2(
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
    v_existing platform.billing_installation_base_assignments%ROWTYPE;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_deployment_mode NOT IN ('onpremise','saas')
       OR p_offering_revision < 1
       OR p_offering_sha256 !~ '^[0-9a-f]{64}$'
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR p_binding_purpose NOT IN ('onpremise_primary','saas_control','saas_customer')
       OR NOT pg_catalog.isfinite(p_offering_effective_at)
       OR (p_deployment_mode='onpremise' AND
           (p_source_sequence < 1 OR p_source_sha256 !~ '^[0-9a-f]{64}$'))
       OR (p_deployment_mode='saas' AND
           (p_source_sequence IS NOT NULL OR p_source_sha256 IS NOT NULL)) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='billing installation assignment input is invalid';
    END IF;
    SELECT * INTO v_existing
    FROM platform.billing_installation_base_assignments AS fact
    WHERE fact.operation_id=p_operation_id;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.tenant_id IS DISTINCT FROM p_tenant_id
           OR v_existing.authority_id IS DISTINCT FROM p_authority_id
           OR v_existing.assignment_id IS DISTINCT FROM p_assignment_id
           OR v_existing.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR v_existing.offering_version_id IS DISTINCT FROM p_offering_version_id
           OR v_existing.offering_revision IS DISTINCT FROM p_offering_revision
           OR v_existing.offering_sha256 IS DISTINCT FROM p_offering_sha256
           OR v_existing.offering_effective_at IS DISTINCT FROM p_offering_effective_at
           OR v_existing.source_sequence IS DISTINCT FROM p_source_sequence
           OR v_existing.source_sha256 IS DISTINCT FROM p_source_sha256
           OR v_existing.manifest_sha256 IS DISTINCT FROM p_manifest_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='billing installation assignment replay conflicts';
        END IF;
        RETURN true;
    END IF;
    IF p_binding_purpose='onpremise_primary' THEN
        PERFORM 1 FROM platform.billing_license_offering_publications AS publication
        WHERE publication.environment_id=p_environment_id
          AND publication.offering_version_id=p_offering_version_id
          AND publication.revision=p_offering_revision
          AND publication.offering_sha256=p_offering_sha256
          AND publication.effective_at=p_offering_effective_at
          AND publication.snapshot_sequence=p_source_sequence
          AND publication.snapshot_sha256=p_source_sha256 FOR SHARE;
    ELSIF p_binding_purpose='saas_control' THEN
        PERFORM 1 FROM platform.billing_saas_control_offering_publications AS publication
        WHERE publication.offering_key='saas_platform_control_base_v1'
          AND publication.offering_version_id=p_offering_version_id
          AND publication.revision=p_offering_revision
          AND publication.offering_sha256=p_offering_sha256
          AND publication.effective_at=p_offering_effective_at FOR SHARE;
    ELSIF p_binding_purpose='saas_customer' THEN
        PERFORM 1 FROM platform.billing_saas_free_base_publications AS publication
        WHERE publication.offering_key='saas_free_base_v1'
          AND publication.offering_version_id=p_offering_version_id
          AND publication.revision=p_offering_revision
          AND publication.offering_sha256=p_offering_sha256
          AND publication.effective_at=p_offering_effective_at FOR SHARE;
    ELSE
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='billing installation purpose is invalid';
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='billing installation offering pin is unavailable';
    END IF;
    PERFORM platform.billing_register_tenant_authority_v1(
        p_authority_id,p_tenant_id,1,1,1,v_now
    );
    INSERT INTO platform.billing_tenant_base_offerings (
        id,authority_id,tenant_id,offering_version_id,effective_start,
        effective_end,source_operation_id,created_at
    ) VALUES (p_assignment_id,p_authority_id,p_tenant_id,p_offering_version_id,
        p_offering_effective_at,NULL,p_operation_id,v_now);
    INSERT INTO platform.billing_installation_base_assignments (
        operation_id,environment_id,tenant_id,authority_id,assignment_id,
        deployment_mode,offering_version_id,offering_revision,offering_sha256,
        offering_effective_at,source_sequence,source_sha256,manifest_sha256,assigned_at
    ) VALUES (p_operation_id,p_environment_id,p_tenant_id,p_authority_id,p_assignment_id,
        p_deployment_mode,p_offering_version_id,p_offering_revision,p_offering_sha256,
        p_offering_effective_at,p_source_sequence,p_source_sha256,p_manifest_sha256,v_now);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_installation_assign_v2(
    uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_assign_v2(
    uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text
) TO harpy_runtime;
