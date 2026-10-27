-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_auth_throttle_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'iam_auth_throttle_buckets' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM throttle session authority is invalid';
    END IF;
    IF TG_OP OPERATOR(pg_catalog.=) 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM throttle deletion is forbidden';
    END IF;

    IF NEW.bucket_scope_kind OPERATOR(pg_catalog.=) 'tenant' THEN
        PERFORM platform.iam_assert_current_tenant_authority_v1(
            NEW.tenant_id,
            NEW.authority_group_generation,
            NEW.cutover_generation
        );
    ELSE
        PERFORM 1
        FROM platform.environment_identity AS environment
        WHERE environment.environment_id
            OPERATOR(pg_catalog.=) NEW.environment_id
          AND NEW.scope_id OPERATOR(pg_catalog.=) NEW.environment_id
        FOR SHARE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = 'P5721',
                MESSAGE = 'IAM throttle environment identity is missing';
        END IF;
    END IF;
    IF TG_OP OPERATOR(pg_catalog.=) 'INSERT' THEN
        RETURN NEW;
    END IF;

    IF NEW.bucket_scope_kind IS DISTINCT FROM OLD.bucket_scope_kind OR
       NEW.scope_id IS DISTINCT FROM OLD.scope_id OR
       NEW.purpose IS DISTINCT FROM OLD.purpose OR
       NEW.dimension IS DISTINCT FROM OLD.dimension OR
       NEW.key_id IS DISTINCT FROM OLD.key_id OR
       NEW.bucket_hmac IS DISTINCT FROM OLD.bucket_hmac OR
       NEW.authority_epoch IS DISTINCT FROM OLD.authority_epoch OR
       NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
       NEW.tenant_id IS DISTINCT FROM OLD.tenant_id OR
       NEW.authority_scope_kind IS DISTINCT FROM OLD.authority_scope_kind OR
       NEW.authority_scope_id IS DISTINCT FROM OLD.authority_scope_id OR
       NEW.authority_group_id IS DISTINCT FROM OLD.authority_group_id OR
       NEW.authority_group_generation IS DISTINCT FROM
           OLD.authority_group_generation OR
       NEW.authority_domain IS DISTINCT FROM OLD.authority_domain OR
       NEW.cutover_generation IS DISTINCT FROM OLD.cutover_generation THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM throttle identity is immutable';
    END IF;

    IF NEW.revision IS DISTINCT FROM
           OLD.revision OPERATOR(pg_catalog.+) 1 OR
       NEW.last_attempt_at OPERATOR(pg_catalog.<=) OLD.last_attempt_at OR
       NEW.retain_until OPERATOR(pg_catalog.<) OLD.retain_until OR
       NEW.policy_version OPERATOR(pg_catalog.<) OLD.policy_version THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM throttle state must advance monotonically';
    END IF;

    IF OLD.blocked_until IS NOT NULL AND
       OLD.blocked_until OPERATOR(pg_catalog.>) NEW.last_attempt_at AND
       (
           NEW.blocked_until IS NULL OR
           NEW.blocked_until OPERATOR(pg_catalog.<) OLD.blocked_until
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM throttle active block cannot shrink';
    END IF;

    IF NEW.window_start IS NOT DISTINCT FROM OLD.window_start AND
       NEW.window_end IS NOT DISTINCT FROM OLD.window_end THEN
        IF NEW.policy_version IS DISTINCT FROM OLD.policy_version OR
           NEW.attempt_count IS DISTINCT FROM
               OLD.attempt_count OPERATOR(pg_catalog.+) 1 OR
           (
               OLD.blocked_until IS NOT NULL AND
               (
                   NEW.blocked_until IS NULL OR
                   NEW.blocked_until OPERATOR(pg_catalog.<) OLD.blocked_until
               )
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'IAM throttle active window update is invalid';
        END IF;
    ELSIF NEW.window_start OPERATOR(pg_catalog.<=) OLD.window_start OR
          (
              NEW.policy_version OPERATOR(pg_catalog.=) OLD.policy_version AND
              NEW.window_start OPERATOR(pg_catalog.<) OLD.window_end
          ) OR
          NEW.attempt_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM throttle window reset is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_auth_throttle_guard_v1() FROM PUBLIC;

CREATE TRIGGER iam_auth_throttle_buckets_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_auth_throttle_buckets
FOR EACH ROW
EXECUTE FUNCTION platform.iam_auth_throttle_guard_v1();

ALTER TABLE platform.iam_auth_throttle_buckets
    ENABLE ALWAYS TRIGGER iam_auth_throttle_buckets_guard;
