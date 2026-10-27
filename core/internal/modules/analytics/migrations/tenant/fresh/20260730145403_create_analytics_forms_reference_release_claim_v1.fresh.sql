-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_forms_reference_active_v1(
    p_scope_schema name,p_source_id uuid,p_keep_version_id uuid,
    p_component_id uuid
)
RETURNS SETOF analytics_forms_reference_bindings
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY SELECT binding.*
    FROM analytics_forms_reference_bindings AS binding
    WHERE (p_source_id IS NULL OR binding.source_id=p_source_id)
      AND binding.registered_at IS NOT NULL
      AND binding.released_at IS NULL
      AND (p_keep_version_id IS NULL OR
           binding.source_version_id<>p_keep_version_id)
      AND (p_component_id IS NULL OR binding.component_id=p_component_id)
      AND (p_source_id IS NOT NULL OR p_component_id IS NOT NULL)
    ORDER BY binding.reference_id;
END;
$function$;
CREATE FUNCTION analytics_forms_reference_release_claim_v1(
    p_scope_schema name,p_reference_id uuid,p_operation_id uuid,
    p_occurred_at timestamptz
)
RETURNS SETOF analytics_forms_reference_bindings
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_existing analytics_forms_reference_bindings%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    IF pg_catalog.uuid_extract_version(p_reference_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_reference_id=p_operation_id OR p_occurred_at IS NULL OR
       NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Forms reference release claim is invalid';
    END IF;
    SELECT * INTO v_existing
    FROM analytics_forms_reference_bindings AS binding
    WHERE binding.reference_id=p_reference_id FOR UPDATE;
    IF NOT FOUND OR v_existing.registered_at IS NULL OR
       p_occurred_at<v_existing.registered_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Forms reference release authority differs';
    END IF;
    IF v_existing.release_operation_id IS NOT NULL THEN
        IF v_existing.release_operation_id<>p_operation_id AND
           v_existing.released_at IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Forms reference release claim differs';
        END IF;
        RETURN NEXT v_existing;
        RETURN;
    END IF;
    UPDATE analytics_forms_reference_bindings AS binding
    SET release_operation_id=p_operation_id,release_claimed_at=p_occurred_at
    WHERE binding.reference_id=p_reference_id
    RETURNING * INTO v_existing;
    RETURN NEXT v_existing;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_forms_reference_active_v1(
    name,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_forms_reference_release_claim_v1(
    name,uuid,uuid,timestamptz
) FROM PUBLIC;
