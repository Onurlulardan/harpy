-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_attempt_v2(
    p_scope_schema name,p_receipt_id uuid,p_projection_receipt_id uuid,
    p_actor_id uuid,p_idempotency_key text,p_request_sha256 bytea,
    p_occurred_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_existing analytics_catalog_projection_attempts%ROWTYPE;
    v_ready_at timestamptz;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_projection_receipt_id)
           IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_idempotency_key IS NULL OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_idempotency_key !~ '^[A-Za-z0-9_.:-]+$' OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection attempt input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=p_projection_receipt_id FOR UPDATE;
    IF NOT FOUND OR p_occurred_at<v_projection.created_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog projection artifact is unavailable';
    END IF;
    SELECT * INTO v_existing
    FROM analytics_catalog_projection_attempts AS attempt
    WHERE attempt.projection_receipt_id=p_projection_receipt_id
      AND attempt.idempotency_key=p_idempotency_key FOR UPDATE;
    IF FOUND THEN
        IF v_existing.actor_id<>p_actor_id OR
           v_existing.request_sha256<>p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Catalog projection attempt differs';
        END IF;
        RETURN v_existing.receipt_id;
    END IF;
    v_ready_at:=CASE WHEN v_projection.operational_state='ready'
        THEN p_occurred_at ELSE NULL END;
    INSERT INTO analytics_catalog_projection_attempts(
        receipt_id,projection_receipt_id,actor_id,idempotency_key,
        request_sha256,ready_at,created_at,updated_at
    ) VALUES (
        p_receipt_id,p_projection_receipt_id,p_actor_id,p_idempotency_key,
        p_request_sha256,v_ready_at,p_occurred_at,p_occurred_at
    );
    RETURN p_receipt_id;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_projection_attempt_v2(
    name,uuid,uuid,uuid,text,bytea,timestamptz
) FROM PUBLIC;
