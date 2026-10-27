-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_platform_owner_user_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_id uuid;
    v_tenant_id uuid;
BEGIN
    v_user_id := OLD.id;
    IF TG_OP = 'UPDATE' AND (
       OLD.status IS NOT DISTINCT FROM NEW.status OR NEW.status = 'active') THEN
        RETURN NEW;
    END IF;

    SELECT identity.tenant_id INTO v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton;

    IF v_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_platform_owner_user_scope_guard',
            MESSAGE = 'IAM platform owner user scope is unavailable';
    END IF;

    IF platform.iam_platform_current_owner_v1(v_tenant_id, v_user_id) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_platform_owner_user_lifecycle_guard',
            MESSAGE = 'Current Instance Owner requires a completed successor transfer';
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_platform_owner_user_guard_v1() FROM PUBLIC;

CREATE TRIGGER iam_users_platform_owner_lifecycle_guard
BEFORE UPDATE OF status OR DELETE ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_platform_owner_user_guard_v1();
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_platform_owner_lifecycle_guard;
