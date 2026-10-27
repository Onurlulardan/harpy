-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(
    p_environment_id uuid,p_control_tenant_id uuid,p_actor_id uuid,
    p_target_tenant_id uuid,p_kind text,p_name text,p_expected_revision bigint,
    p_idempotency_key text,p_request_sha256 bytea,p_now timestamptz
)
RETURNS TABLE (
    outcome text,replay boolean,id uuid,name text,slug text,status text,
    lifecycle_revision bigint,settings_revision bigint,
    created_at timestamptz,updated_at timestamptz,
    offboarded_at timestamptz,purged_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_tenant platform.tenants%ROWTYPE;
    v_digest bytea;
    v_kind text;
    v_expected bigint;
    v_document jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_environment_id IS NULL OR p_control_tenant_id IS NULL OR
       p_actor_id IS NULL OR p_target_tenant_id IS NULL OR
       p_target_tenant_id=p_control_tenant_id OR
       p_kind NOT IN ('update','suspend','reactivate') OR
       p_expected_revision < 1 OR p_idempotency_key IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       p_request_sha256 IS NULL OR pg_catalog.length(p_request_sha256)<>32 OR
       p_now IS NULL OR
       pg_catalog.abs(EXTRACT(EPOCH FROM
           (pg_catalog.transaction_timestamp()-p_now))) > 300 OR
       (p_kind='update' AND (p_name IS NULL OR p_name<>pg_catalog.btrim(p_name) OR
           pg_catalog.char_length(p_name) NOT BETWEEN 1 AND 255)) OR
       (p_kind<>'update' AND p_name IS NOT NULL) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Customer tenant lifecycle mutation invalid';
    END IF;
    PERFORM 1 FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS control_binding
      ON control_binding.environment_id=policy.environment_id AND
         control_binding.tenant_id=p_control_tenant_id AND
         control_binding.purpose='saas_control'
    WHERE policy.singleton AND policy.environment_id=p_environment_id AND
          policy.deployment_mode='saas';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='SaaS control tenant boundary is required';
    END IF;
    SELECT receipt.request_sha256,receipt.kind,receipt.expected_revision,
           receipt.result_document
      INTO v_digest,v_kind,v_expected,v_document
    FROM platform.tenancy_customer_lifecycle_receipts AS receipt
    WHERE receipt.tenant_id=p_target_tenant_id AND
          receipt.idempotency_key=p_idempotency_key;
    IF FOUND THEN
        IF v_digest<>p_request_sha256 OR v_kind<>p_kind OR
           v_expected<>p_expected_revision THEN
            RETURN QUERY SELECT 'idempotency_conflict',false,NULL::uuid,
                NULL::text,NULL::text,NULL::text,NULL::bigint,NULL::bigint,
                NULL::timestamptz,NULL::timestamptz,NULL::timestamptz,NULL::timestamptz;
        ELSE
            RETURN QUERY SELECT 'replay',true,(v_document->>'id')::uuid,
                v_document->>'name',v_document->>'slug',v_document->>'status',
                (v_document->>'revision')::bigint,
                (v_document->>'settings_revision')::bigint,
                (v_document->>'created_at')::timestamptz,
                (v_document->>'updated_at')::timestamptz,
                (v_document->>'offboarded_at')::timestamptz,
                (v_document->>'purged_at')::timestamptz;
        END IF;
        RETURN;
    END IF;
    SELECT tenant.* INTO v_tenant
    FROM platform.tenancy_tenant_bindings AS binding
    JOIN platform.tenants AS tenant ON tenant.id=binding.tenant_id
    WHERE binding.environment_id=p_environment_id AND
          binding.tenant_id=p_target_tenant_id AND binding.purpose='saas_customer'
    FOR UPDATE OF tenant;
    IF NOT FOUND THEN
        RETURN QUERY SELECT 'not_found',false,NULL::uuid,NULL::text,NULL::text,
            NULL::text,NULL::bigint,NULL::bigint,NULL::timestamptz,
            NULL::timestamptz,NULL::timestamptz,NULL::timestamptz;
        RETURN;
    END IF;
    IF v_tenant.lifecycle_revision<>p_expected_revision THEN
        RETURN QUERY SELECT 'revision_conflict',false,NULL::uuid,NULL::text,
            NULL::text,NULL::text,NULL::bigint,NULL::bigint,NULL::timestamptz,
            NULL::timestamptz,NULL::timestamptz,NULL::timestamptz;
        RETURN;
    END IF;
    IF v_tenant.deleted_at IS NOT NULL OR v_tenant.purged_at IS NOT NULL OR
       (p_kind='suspend' AND v_tenant.status NOT IN ('ACTIVE','TRIAL')) OR
       (p_kind='reactivate' AND v_tenant.status<>'SUSPENDED') THEN
        RETURN QUERY SELECT 'state_conflict',false,NULL::uuid,NULL::text,
            NULL::text,NULL::text,NULL::bigint,NULL::bigint,NULL::timestamptz,
            NULL::timestamptz,NULL::timestamptz,NULL::timestamptz;
        RETURN;
    END IF;
    UPDATE platform.tenants AS tenant SET
        name=CASE WHEN p_kind='update' THEN p_name ELSE tenant.name END,
        status=CASE p_kind WHEN 'suspend' THEN 'SUSPENDED'
                           WHEN 'reactivate' THEN 'ACTIVE' ELSE tenant.status END,
        lifecycle_revision=tenant.lifecycle_revision+1,
        settings_revision=tenant.settings_revision+
            CASE WHEN p_kind='update' THEN 1 ELSE 0 END,
        updated_at=p_now
    WHERE tenant.id=p_target_tenant_id RETURNING tenant.* INTO v_tenant;
    v_document:=pg_catalog.jsonb_build_object(
        'id',v_tenant.id,'name',v_tenant.name,'slug',v_tenant.slug,
        'status',v_tenant.status,'revision',v_tenant.lifecycle_revision,
        'settings_revision',v_tenant.settings_revision,
        'created_at',v_tenant.created_at,'updated_at',v_tenant.updated_at,
        'offboarded_at',v_tenant.offboarded_at,'purged_at',v_tenant.purged_at);
    INSERT INTO platform.tenancy_customer_lifecycle_receipts(
        tenant_id,control_tenant_id,actor_id,kind,idempotency_key,request_sha256,
        expected_revision,result_revision,result_document,created_at
    ) VALUES (p_target_tenant_id,p_control_tenant_id,p_actor_id,p_kind,
        p_idempotency_key,p_request_sha256,p_expected_revision,
        v_tenant.lifecycle_revision,v_document,p_now);
    RETURN QUERY SELECT 'applied',false,v_tenant.id,v_tenant.name::text,
        v_tenant.slug::text,v_tenant.status::text,v_tenant.lifecycle_revision,
        v_tenant.settings_revision,v_tenant.created_at,v_tenant.updated_at,
        v_tenant.offboarded_at,v_tenant.purged_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(
    uuid,uuid,uuid,uuid,text,text,bigint,text,bytea,timestamptz
) FROM PUBLIC;
