-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_forms_reference_register_ack_v1(
    p_scope_schema name,p_reference_id uuid,p_operation_id uuid,
    p_target_epoch bigint,p_result_sha256 bytea,p_registered_at timestamptz
)
RETURNS SETOF analytics_forms_reference_bindings
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_forms_reference_bindings%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    IF pg_catalog.uuid_extract_version(p_reference_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_target_epoch IS NULL OR p_target_epoch<1 OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       p_registered_at IS NULL OR NOT pg_catalog.isfinite(p_registered_at) OR
       date_trunc('microseconds',p_registered_at)<>p_registered_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Forms reference register ack is invalid';
    END IF;
    SELECT * INTO v_existing
    FROM analytics_forms_reference_bindings AS binding
    WHERE binding.reference_id=p_reference_id FOR UPDATE;
    IF NOT FOUND OR v_existing.register_operation_id<>p_operation_id OR
       v_existing.claimed_at<>p_registered_at OR
       v_existing.release_operation_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Forms reference register authority differs';
    END IF;
    IF v_existing.registered_at IS NOT NULL THEN
        IF v_existing.target_reference_epoch<>p_target_epoch OR
           v_existing.register_result_sha256<>p_result_sha256 OR
           v_existing.registered_at<>p_registered_at THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Forms reference register ack differs';
        END IF;
        RETURN NEXT v_existing;
        RETURN;
    END IF;
    UPDATE analytics_forms_reference_bindings AS binding
    SET target_reference_epoch=p_target_epoch,
        register_result_sha256=p_result_sha256,
        registered_at=p_registered_at
    WHERE binding.reference_id=p_reference_id
    RETURNING * INTO v_existing;
    RETURN NEXT v_existing;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_forms_reference_register_ack_v1(
    name,uuid,uuid,bigint,bytea,timestamptz
) FROM PUBLIC;
