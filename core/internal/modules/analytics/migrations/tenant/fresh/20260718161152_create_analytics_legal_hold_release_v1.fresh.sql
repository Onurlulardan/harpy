-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_legal_hold_release_v1(
    p_scope_schema name,p_hold_id uuid,p_actor_id uuid,
    p_expected_revision bigint,p_released_at timestamptz
)
RETURNS TABLE (hold_id uuid,state text,row_revision bigint,replayed boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_hold analytics_legal_holds%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_hold_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_expected_revision < 1 OR p_released_at IS NULL OR
       NOT pg_catalog.isfinite(p_released_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics legal hold release is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT hold_row.* INTO v_hold FROM analytics_legal_holds AS hold_row
    WHERE hold_row.hold_id = p_hold_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics legal hold was not found';
    END IF;
    IF v_hold.state = 'released' THEN
        IF v_hold.released_by IS DISTINCT FROM p_actor_id OR
           v_hold.row_revision IS DISTINCT FROM p_expected_revision + 1 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103',
                MESSAGE = 'analytics legal hold release differs';
        END IF;
        RETURN QUERY SELECT v_hold.hold_id,v_hold.state,v_hold.row_revision,true;
        RETURN;
    END IF;
    IF v_hold.row_revision IS DISTINCT FROM p_expected_revision OR
       p_released_at < v_hold.placed_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'analytics legal hold revision changed';
    END IF;
    UPDATE analytics_legal_holds
    SET state = 'released',row_revision = analytics_legal_holds.row_revision + 1,
        released_by = p_actor_id,released_at = p_released_at
    WHERE analytics_legal_holds.hold_id = p_hold_id;
    RETURN QUERY SELECT p_hold_id,'released'::text,p_expected_revision + 1,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_legal_hold_release_v1(
    name,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_legal_hold_release_v1(
    name,uuid,uuid,bigint,timestamptz
) TO harpy_runtime;
