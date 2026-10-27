-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource_kind text,
    p_operation_id uuid,
    p_owner_fact_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_evidence_sha256 bytea,
    p_evidence_exists boolean,
    p_evidence_active boolean,
    p_included_in_opening boolean,
    p_changed boolean,
    p_gauge_delta bigint,
    p_current_amount bigint,
    p_limit_amount bigint,
    p_unlimited boolean,
    p_usage_decision_id uuid,
    p_usage_entry_id uuid,
    p_before_revision bigint,
    p_recorded_at timestamptz,
    p_measured_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_gauge_key text;
    v_want_state text;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    v_want_state := CASE
        WHEN p_evidence_active THEN 'active'
        ELSE 'released'
    END;
    v_gauge_key := CASE p_resource_kind
        WHEN 'api_endpoint' THEN 'api.active_endpoints'
        WHEN 'analytics_dashboard' THEN 'analytics.active_dashboards'
    END;
    IF v_gauge_key IS NULL OR p_actor_id IS NULL
       OR pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_owner_fact_id) IS DISTINCT FROM 7
       OR p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$'
       OR pg_catalog.octet_length(p_request_sha256) <> 32
       OR pg_catalog.octet_length(p_evidence_sha256) <> 32
       OR (NOT p_evidence_exists AND p_evidence_active)
       OR (p_included_in_opening
           AND (NOT p_evidence_exists OR NOT p_evidence_active))
       OR p_gauge_delta NOT BETWEEN -1 AND 1
       OR (NOT p_changed AND p_gauge_delta <> 0)
       OR p_current_amount < 0 OR p_limit_amount < 0
       OR (p_unlimited AND p_limit_amount <> 0)
       OR p_before_revision < 0
       OR NOT pg_catalog.isfinite(p_recorded_at)
       OR NOT pg_catalog.isfinite(p_measured_at)
       OR NOT EXISTS (
           SELECT 1 FROM platform.billing_tenant_authorities AS authority_row
           WHERE authority_row.authority_id=p_authority_id
             AND authority_row.tenant_id=p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing metered resource reconciliation record is invalid';
    END IF;

    IF p_gauge_delta = 1 AND NOT platform.billing_metered_resource_feature_v1(
        p_authority_id,p_tenant_id,
        CASE p_resource_kind
            WHEN 'api_endpoint' THEN 'api.builder.enabled'
            ELSE 'analytics.dashboard_builder.enabled'
        END
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing metered resource builder is unavailable';
    END IF;

    IF p_gauge_delta <> 0 THEN
        IF p_usage_decision_id IS NULL OR p_usage_entry_id IS NULL
           OR NOT EXISTS (
               SELECT 1
               FROM platform.billing_entitlement_decisions AS decision_row
               JOIN platform.billing_usage_entries AS entry_row
                 ON entry_row.decision_id=decision_row.id
               WHERE decision_row.id=p_usage_decision_id
                 AND decision_row.authority_id=p_authority_id
                 AND decision_row.tenant_id=p_tenant_id
                 AND decision_row.actor_id=p_actor_id
                 AND decision_row.entitlement_key=v_gauge_key
                 AND decision_row.source_operation_id=p_operation_id
                 AND decision_row.requested_amount=p_gauge_delta
                 AND decision_row.current_amount=p_current_amount
                 AND decision_row.unlimited=p_unlimited
                 AND (
                     (p_unlimited AND decision_row.limit_amount IS NULL)
                     OR (NOT p_unlimited
                         AND decision_row.limit_amount=p_limit_amount)
                 )
                 AND decision_row.allowed
                 AND entry_row.id=p_usage_entry_id
                 AND entry_row.authority_id=p_authority_id
                 AND entry_row.tenant_id=p_tenant_id
                 AND entry_row.entitlement_key=v_gauge_key
                 AND entry_row.owner_fact_id=p_operation_id
                 AND entry_row.entry_kind='gauge_delta'
                 AND entry_row.amount_delta=p_gauge_delta
                 AND entry_row.gauge_after=p_current_amount+p_gauge_delta
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing reconciliation usage evidence is unavailable';
        END IF;
    ELSIF p_usage_decision_id IS NOT NULL OR p_usage_entry_id IS NOT NULL
          OR p_current_amount <> 0 OR p_limit_amount <> 0 OR p_unlimited THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing reconciliation usage evidence is unexpected';
    END IF;

    IF NOT p_changed THEN
        IF p_before_revision <> 0 OR NOT EXISTS (
            SELECT 1 FROM platform.billing_metered_resources AS projection_row
            WHERE projection_row.authority_id=p_authority_id
              AND projection_row.tenant_id=p_tenant_id
              AND projection_row.resource_kind=p_resource_kind
              AND projection_row.owner_fact_id=p_owner_fact_id
              AND projection_row.counted_state=v_want_state
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing reconciliation projection is unavailable';
        END IF;
    ELSIF p_before_revision = 0 THEN
        IF p_gauge_delta <> 0
           OR (p_evidence_active AND NOT p_included_in_opening) THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'billing reconciliation adoption evidence is invalid';
        END IF;
        INSERT INTO platform.billing_metered_resources (
            authority_id,tenant_id,resource_kind,owner_fact_id,counted_state,
            origin_kind,opening_evidence_sha256,opening_measured_at,
            last_operation_id,revision,created_at,updated_at
        ) VALUES (
            p_authority_id,p_tenant_id,p_resource_kind,p_owner_fact_id,
            v_want_state,'adopted',p_evidence_sha256,p_measured_at,
            p_operation_id,1,p_recorded_at,p_recorded_at
        );
    ELSE
        IF (p_evidence_active AND p_gauge_delta <> 1)
           OR (NOT p_evidence_active AND p_gauge_delta <> -1) THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'billing reconciliation gauge direction is invalid';
        END IF;
        UPDATE platform.billing_metered_resources AS projection_row
        SET counted_state=v_want_state,last_operation_id=p_operation_id,
            revision=projection_row.revision+1,updated_at=p_recorded_at
        WHERE projection_row.authority_id=p_authority_id
          AND projection_row.tenant_id=p_tenant_id
          AND projection_row.resource_kind=p_resource_kind
          AND projection_row.owner_fact_id=p_owner_fact_id
          AND projection_row.counted_state<>v_want_state
          AND projection_row.revision=p_before_revision;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing reconciliation projection revision conflicts';
        END IF;
    END IF;

    INSERT INTO platform.billing_metered_resource_reconcile_receipts (
        operation_id,authority_id,tenant_id,actor_id,resource_kind,owner_fact_id,
        idempotency_key,request_sha256,evidence_sha256,evidence_exists,
        evidence_active,included_in_opening,changed,gauge_delta,current_amount,
        limit_amount,unlimited,usage_decision_id,usage_entry_id,measured_at,
        recorded_at
    ) VALUES (
        p_operation_id,p_authority_id,p_tenant_id,p_actor_id,p_resource_kind,
        p_owner_fact_id,p_idempotency_key,p_request_sha256,p_evidence_sha256,
        p_evidence_exists,p_evidence_active,p_included_in_opening,p_changed,
        p_gauge_delta,p_current_amount,p_limit_amount,p_unlimited,
        p_usage_decision_id,p_usage_entry_id,p_measured_at,p_recorded_at
    );
    RETURN true;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    uuid,uuid,uuid,text,uuid,uuid,text,bytea,bytea,boolean,boolean,boolean,
    boolean,bigint,bigint,bigint,boolean,uuid,uuid,bigint,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    uuid,uuid,uuid,text,uuid,uuid,text,bytea,bytea,boolean,boolean,boolean,
    boolean,bigint,bigint,bigint,boolean,uuid,uuid,bigint,timestamptz,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
