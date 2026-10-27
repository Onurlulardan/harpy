-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_record_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource_kind text,
    p_mutation_kind text,
    p_operation_id uuid,
    p_owner_fact_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_feature_enabled boolean,
    p_allowed boolean,
    p_gauge_delta bigint,
    p_current_amount bigint,
    p_limit_amount bigint,
    p_unlimited boolean,
    p_usage_decision_id uuid,
    p_usage_entry_id uuid,
    p_before_revision bigint,
    p_recorded_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_gauge_key text;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    v_gauge_key := CASE p_resource_kind
        WHEN 'api_endpoint' THEN 'api.active_endpoints'
        WHEN 'analytics_dashboard' THEN 'analytics.active_dashboards'
    END;
    IF v_gauge_key IS NULL OR p_mutation_kind NOT IN (
           'create','update','publish','revalidate','delete','restore'
       )
       OR pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_owner_fact_id) IS DISTINCT FROM 7
       OR p_actor_id IS NULL
       OR p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$'
       OR pg_catalog.octet_length(p_request_sha256) <> 32
       OR p_current_amount < 0 OR p_limit_amount < 0
       OR (p_unlimited AND p_limit_amount <> 0)
       OR p_before_revision < 0 OR NOT pg_catalog.isfinite(p_recorded_at)
       OR NOT EXISTS (
           SELECT 1 FROM platform.billing_tenant_authorities AS authority_row
           WHERE authority_row.authority_id = p_authority_id
             AND authority_row.tenant_id = p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing metered resource record input is invalid';
    END IF;

    IF p_gauge_delta <> 0 OR (
        p_mutation_kind IN ('create','restore')
        AND p_feature_enabled AND NOT p_allowed
    ) THEN
        IF p_usage_decision_id IS NULL OR NOT EXISTS (
            SELECT 1
            FROM platform.billing_entitlement_decisions AS decision_row
            WHERE decision_row.id = p_usage_decision_id
              AND decision_row.authority_id = p_authority_id
              AND decision_row.tenant_id = p_tenant_id
              AND decision_row.actor_id = p_actor_id
              AND decision_row.entitlement_key = v_gauge_key
              AND decision_row.source_operation_id = p_operation_id
              AND decision_row.requested_amount = CASE
                  WHEN p_gauge_delta <> 0 THEN p_gauge_delta ELSE 1
              END
              AND decision_row.current_amount = p_current_amount
              AND decision_row.unlimited = p_unlimited
              AND (
                  (p_unlimited AND decision_row.limit_amount IS NULL)
                  OR (NOT p_unlimited
                      AND decision_row.limit_amount = p_limit_amount)
              )
              AND decision_row.allowed = p_allowed
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing usage decision evidence is unavailable';
        END IF;
    ELSIF p_usage_decision_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing usage decision evidence is unexpected';
    END IF;

    IF p_gauge_delta <> 0 THEN
        IF p_usage_entry_id IS NULL OR NOT EXISTS (
            SELECT 1
            FROM platform.billing_usage_entries AS entry_row
            WHERE entry_row.id = p_usage_entry_id
              AND entry_row.authority_id = p_authority_id
              AND entry_row.tenant_id = p_tenant_id
              AND entry_row.decision_id = p_usage_decision_id
              AND entry_row.entitlement_key = v_gauge_key
              AND entry_row.owner_fact_id = p_operation_id
              AND entry_row.entry_kind = 'gauge_delta'
              AND entry_row.amount_delta = p_gauge_delta
              AND entry_row.gauge_after = p_current_amount + p_gauge_delta
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing usage entry evidence is unavailable';
        END IF;
    ELSIF p_usage_entry_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing usage entry evidence is unexpected';
    END IF;

    IF p_mutation_kind = 'create' AND p_allowed THEN
        IF p_before_revision <> 0 OR p_gauge_delta <> 1 THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'billing metered resource create transition is invalid';
        END IF;
        INSERT INTO platform.billing_metered_resources (
            authority_id,tenant_id,resource_kind,owner_fact_id,counted_state,
            origin_kind,opening_evidence_sha256,opening_measured_at,
            last_operation_id,revision,created_at,updated_at
        ) VALUES (
            p_authority_id,p_tenant_id,p_resource_kind,p_owner_fact_id,'active',
            'runtime',NULL,NULL,p_operation_id,1,p_recorded_at,p_recorded_at
        );
    ELSIF p_mutation_kind = 'restore' AND p_allowed THEN
        UPDATE platform.billing_metered_resources AS projection_row
        SET counted_state='active',last_operation_id=p_operation_id,
            revision=projection_row.revision+1,updated_at=p_recorded_at
        WHERE projection_row.authority_id=p_authority_id
          AND projection_row.tenant_id=p_tenant_id
          AND projection_row.resource_kind=p_resource_kind
          AND projection_row.owner_fact_id=p_owner_fact_id
          AND projection_row.counted_state='released'
          AND projection_row.revision=p_before_revision;
        IF NOT FOUND OR p_gauge_delta <> 1 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing metered resource restore revision conflicts';
        END IF;
    ELSIF p_mutation_kind = 'delete' AND p_gauge_delta = -1 THEN
        UPDATE platform.billing_metered_resources AS projection_row
        SET counted_state='released',last_operation_id=p_operation_id,
            revision=projection_row.revision+1,updated_at=p_recorded_at
        WHERE projection_row.authority_id=p_authority_id
          AND projection_row.tenant_id=p_tenant_id
          AND projection_row.resource_kind=p_resource_kind
          AND projection_row.owner_fact_id=p_owner_fact_id
          AND projection_row.counted_state='active'
          AND projection_row.revision=p_before_revision;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing metered resource delete revision conflicts';
        END IF;
    ELSIF p_mutation_kind = 'delete' THEN
        IF p_gauge_delta <> 0 OR NOT EXISTS (
            SELECT 1 FROM platform.billing_metered_resources AS projection_row
            WHERE projection_row.authority_id=p_authority_id
              AND projection_row.tenant_id=p_tenant_id
              AND projection_row.resource_kind=p_resource_kind
              AND projection_row.owner_fact_id=p_owner_fact_id
              AND projection_row.counted_state='released'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing metered resource release state is unavailable';
        END IF;
    END IF;

    INSERT INTO platform.billing_metered_resource_receipts (
        operation_id,authority_id,tenant_id,actor_id,resource_kind,mutation_kind,
        owner_fact_id,idempotency_key,request_sha256,feature_enabled,allowed,
        gauge_delta,current_amount,limit_amount,unlimited,usage_decision_id,
        usage_entry_id,recorded_at
    ) VALUES (
        p_operation_id,p_authority_id,p_tenant_id,p_actor_id,p_resource_kind,
        p_mutation_kind,p_owner_fact_id,p_idempotency_key,p_request_sha256,
        p_feature_enabled,p_allowed,p_gauge_delta,p_current_amount,p_limit_amount,
        p_unlimited,p_usage_decision_id,p_usage_entry_id,p_recorded_at
    );
    RETURN true;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_metered_resource_record_v1(
    uuid,uuid,uuid,text,text,uuid,uuid,text,bytea,boolean,boolean,bigint,bigint,
    bigint,boolean,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_record_v1(
    uuid,uuid,uuid,text,text,uuid,uuid,text,bytea,boolean,boolean,bigint,bigint,
    bigint,boolean,uuid,uuid,bigint,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
