-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_action_v1(
    p_action_key text, p_occurred_at timestamptz
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_action iam_actions%ROWTYPE;
BEGIN
    SELECT * INTO v_action FROM iam_actions AS action
    WHERE action.action_key=p_action_key FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO iam_actions(
            id,action_key,display_name,description,is_protected,
            state,revision,created_at,updated_at
        ) VALUES (
            pg_catalog.uuidv7(),p_action_key,p_action_key,
            'First-party module catalog action',false,
            'active',1,p_occurred_at,p_occurred_at
        ) RETURNING * INTO v_action;
    ELSIF v_action.state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog operational action differs';
    END IF;
    RETURN v_action.id;
END;
$function$;

CREATE FUNCTION iam_catalog_operational_permission_v1(
    p_resource_id uuid, p_action_id uuid, p_occurred_at timestamptz
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_permission iam_permissions%ROWTYPE;
BEGIN
    SELECT * INTO v_permission FROM iam_permissions AS permission
    WHERE permission.match_kind='exact'
      AND permission.resource_id=p_resource_id
      AND permission.action_id=p_action_id
    FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO iam_permissions(
            id,match_kind,resource_id,action_id,is_protected,
            state,revision,created_at,updated_at
        ) VALUES (
            pg_catalog.uuidv7(),'exact',p_resource_id,p_action_id,
            false,'active',1,p_occurred_at,p_occurred_at
        ) RETURNING * INTO v_permission;
    ELSIF v_permission.state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog operational permission differs';
    END IF;
    RETURN v_permission.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_operational_action_v1(
    text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_catalog_operational_permission_v1(
    uuid,uuid,timestamptz
) FROM PUBLIC;
