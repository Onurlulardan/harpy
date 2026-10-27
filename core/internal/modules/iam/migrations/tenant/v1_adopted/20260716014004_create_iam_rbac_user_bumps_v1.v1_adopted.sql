-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_bump_all_users_v1(p_clock bigint)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_count bigint;
BEGIN
    UPDATE iam_users SET
        authorization_version = CASE WHEN authorization_version + 1 > p_clock
            THEN authorization_version + 1 ELSE p_clock END,
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE status = 'active';
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;

CREATE FUNCTION iam_rbac_bump_user_v1(p_user_id uuid, p_clock bigint)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_count bigint;
BEGIN
    UPDATE iam_users SET
        authorization_version = CASE WHEN authorization_version + 1 > p_clock
            THEN authorization_version + 1 ELSE p_clock END,
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_user_id AND status = 'active';
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002', MESSAGE = 'IAM RBAC user is unavailable';
    END IF;
    RETURN v_count;
END;
$function$;

CREATE FUNCTION iam_rbac_bump_role_users_v1(p_role_id uuid, p_clock bigint)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_count bigint;
BEGIN
    UPDATE iam_users AS principal SET
        authorization_version = CASE WHEN principal.authorization_version + 1 > p_clock
            THEN principal.authorization_version + 1 ELSE p_clock END,
        revision = principal.revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE principal.status = 'active' AND principal.id IN (
        SELECT assignment.user_id FROM iam_user_role_assignments AS assignment
        WHERE assignment.role_id = p_role_id AND assignment.state = 'active'
    );
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;

CREATE FUNCTION iam_rbac_bump_organization_users_v1(
    p_organization_id uuid, p_clock bigint
)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_count bigint;
BEGIN
    UPDATE iam_users AS principal SET
        authorization_version = CASE WHEN principal.authorization_version + 1 > p_clock
            THEN principal.authorization_version + 1 ELSE p_clock END,
        revision = principal.revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE principal.status = 'active' AND principal.id IN (
        SELECT membership.user_id FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = p_organization_id AND membership.state = 'active'
    );
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_bump_all_users_v1(bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_bump_user_v1(uuid,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_bump_role_users_v1(uuid,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_bump_organization_users_v1(uuid,bigint) FROM PUBLIC;
