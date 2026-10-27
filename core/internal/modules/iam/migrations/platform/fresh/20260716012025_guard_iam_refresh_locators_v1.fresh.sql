-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_refresh_locator_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'iam_refresh_locators' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh locator session authority is invalid';
    END IF;
    IF TG_OP OPERATOR(pg_catalog.=) 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM refresh locator deletion is forbidden';
    END IF;

    PERFORM platform.iam_assert_current_tenant_authority_v1(
        NEW.tenant_id,
        NEW.authority_group_generation,
        NEW.cutover_generation
    );
    IF TG_OP OPERATOR(pg_catalog.=) 'INSERT' THEN
        RETURN NEW;
    END IF;

    IF NEW.selector IS DISTINCT FROM OLD.selector OR
       NEW.tenant_id IS DISTINCT FROM OLD.tenant_id OR
       NEW.token_id IS DISTINCT FROM OLD.token_id OR
       NEW.family_id IS DISTINCT FROM OLD.family_id OR
       NEW.scope_kind IS DISTINCT FROM OLD.scope_kind OR
       NEW.scope_id IS DISTINCT FROM OLD.scope_id OR
       NEW.authority_group_id IS DISTINCT FROM OLD.authority_group_id OR
       NEW.authority_group_generation IS DISTINCT FROM
           OLD.authority_group_generation OR
       NEW.authority_domain IS DISTINCT FROM OLD.authority_domain OR
       NEW.cutover_generation IS DISTINCT FROM OLD.cutover_generation OR
       NEW.issued_at IS DISTINCT FROM OLD.issued_at OR
       NEW.family_absolute_expires_at IS DISTINCT FROM
           OLD.family_absolute_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM refresh locator identity is immutable';
    END IF;

    IF NEW.source_revision OPERATOR(pg_catalog.<=) OLD.source_revision OR
       NEW.source_sha256 IS NOT DISTINCT FROM OLD.source_sha256 OR
       NEW.projected_at OPERATOR(pg_catalog.<) OLD.projected_at OR
       NEW.retain_until OPERATOR(pg_catalog.<) OLD.retain_until OR
       (
           OLD.source_token_state IS DISTINCT FROM 'active' AND
           NEW.source_token_state IS DISTINCT FROM OLD.source_token_state
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM refresh locator projection must advance monotonically';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_refresh_locator_guard_v1() FROM PUBLIC;

CREATE TRIGGER iam_refresh_locators_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_refresh_locators
FOR EACH ROW
EXECUTE FUNCTION platform.iam_refresh_locator_guard_v1();

ALTER TABLE platform.iam_refresh_locators
    ENABLE ALWAYS TRIGGER iam_refresh_locators_guard;
