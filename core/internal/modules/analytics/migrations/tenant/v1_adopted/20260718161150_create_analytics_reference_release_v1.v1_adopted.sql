-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_reference_release_v1(
    p_scope_schema name,p_reference_id uuid,p_actor_id uuid,
    p_released_at timestamptz
)
RETURNS TABLE (reference_id uuid,state text,released_at timestamptz,replayed boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_reference analytics_inbound_references%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_reference_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_released_at IS NULL OR NOT pg_catalog.isfinite(p_released_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics reference release is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT reference_row.* INTO v_reference
    FROM analytics_inbound_references AS reference_row
    WHERE reference_row.reference_id = p_reference_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics reference was not found';
    END IF;
    IF v_reference.state = 'released' THEN
        IF v_reference.released_by IS DISTINCT FROM p_actor_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103',
                MESSAGE = 'analytics reference release actor differs';
        END IF;
        RETURN QUERY SELECT v_reference.reference_id,v_reference.state,
            v_reference.released_at,true;
        RETURN;
    END IF;
    IF p_released_at < v_reference.created_at THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics reference release time is invalid';
    END IF;
    UPDATE analytics_inbound_references
    SET state = 'released',released_by = p_actor_id,released_at = p_released_at
    WHERE analytics_inbound_references.reference_id = p_reference_id;
    RETURN QUERY SELECT p_reference_id,'released'::text,p_released_at,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_reference_release_v1(
    name,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_reference_release_v1(
    name,uuid,uuid,timestamptz
) TO harpy_runtime;
