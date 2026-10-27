-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_legal_hold_place_v1(
    p_scope_schema name,p_hold_id uuid,p_actor_id uuid,
    p_reason_sha256 bytea,p_placed_at timestamptz
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
       pg_catalog.octet_length(p_reason_sha256) IS DISTINCT FROM 32 OR
       p_placed_at IS NULL OR NOT pg_catalog.isfinite(p_placed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics legal hold placement is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:legal-hold',6100));
    SELECT hold_row.* INTO v_hold FROM analytics_legal_holds AS hold_row
    WHERE hold_row.hold_id = p_hold_id FOR UPDATE;
    IF FOUND THEN
        IF v_hold.reason_sha256 IS DISTINCT FROM p_reason_sha256 OR
           v_hold.placed_by IS DISTINCT FROM p_actor_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103',
                MESSAGE = 'analytics legal hold identity was reused';
        END IF;
        RETURN QUERY SELECT v_hold.hold_id,v_hold.state,v_hold.row_revision,true;
        RETURN;
    END IF;
    BEGIN
        INSERT INTO analytics_legal_holds (
            hold_id,reason_sha256,placed_by,placed_at,state,row_revision
        ) VALUES (p_hold_id,p_reason_sha256,p_actor_id,p_placed_at,'active',1);
    EXCEPTION WHEN unique_violation THEN
        RAISE EXCEPTION USING ERRCODE = 'P6103',
            MESSAGE = 'analytics tenant already has an active legal hold';
    END;
    RETURN QUERY SELECT p_hold_id,'active'::text,1::bigint,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_legal_hold_place_v1(
    name,uuid,uuid,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_legal_hold_place_v1(
    name,uuid,uuid,bytea,timestamptz
) TO forgecrud_runtime;
