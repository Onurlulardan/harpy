-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_delivery_retry_v1(
    p_operation_id uuid,p_lease_owner_id uuid,p_lease_generation bigint,
    p_error_code text,p_permanent boolean
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_job platform.platformauthority_delivery_jobs%ROWTYPE;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_error_code !~ '^[A-Z][A-Z0-9_]{0,127}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='protected authority delivery error code is invalid';
    END IF;
    SELECT * INTO v_job FROM platform.platformauthority_delivery_jobs AS job
     WHERE job.operation_id=p_operation_id FOR UPDATE;
    IF NOT FOUND OR v_job.state<>'running' OR
       v_job.lease_owner_id<>p_lease_owner_id OR
       v_job.lease_generation<>p_lease_generation OR v_job.lease_until<=v_now THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='protected authority delivery lease is stale';
    END IF;
    UPDATE platform.platformauthority_delivery_jobs
       SET state=CASE WHEN p_permanent OR attempts>=max_attempts
           THEN 'dead' ELSE 'queued' END,
           lease_owner_id=NULL,lease_until=NULL,last_error_code=p_error_code,
           next_attempt_at=CASE WHEN p_permanent OR attempts>=max_attempts
               THEN next_attempt_at
               ELSE v_now+pg_catalog.make_interval(
                   secs=>LEAST(300,1 << (attempts-1))) END,
           updated_at=v_now
     WHERE operation_id=p_operation_id;
END;
$function$;

CREATE FUNCTION platform.platformauthority_delivery_await_import_v1(
    p_operation_id uuid,p_lease_owner_id uuid,p_lease_generation bigint
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    UPDATE platform.platformauthority_delivery_jobs AS job
       SET state='awaiting_import',lease_owner_id=NULL,lease_until=NULL,
           last_error_code=NULL,updated_at=v_now
     WHERE job.operation_id=p_operation_id AND job.state='running'
       AND job.lease_owner_id=p_lease_owner_id
       AND job.lease_generation=p_lease_generation AND job.lease_until>v_now
       AND EXISTS (
           SELECT 1 FROM platform.platformauthority_operations AS authority
            WHERE authority.operation_id=job.operation_id
              AND authority.status<>'completed'
       );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='protected authority delivery lease is stale';
    END IF;
END;
$function$;

CREATE FUNCTION platform.platformauthority_delivery_succeed_v1(
    p_operation_id uuid,p_lease_owner_id uuid,p_lease_generation bigint
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    UPDATE platform.platformauthority_delivery_jobs AS job
       SET state='succeeded',lease_owner_id=NULL,lease_until=NULL,
           last_error_code=NULL,completed_at=v_now,updated_at=v_now
     WHERE job.operation_id=p_operation_id AND job.state='running'
       AND job.lease_owner_id=p_lease_owner_id
       AND job.lease_generation=p_lease_generation AND job.lease_until>v_now
       AND EXISTS (
           SELECT 1 FROM platform.platformauthority_operations AS authority
            WHERE authority.operation_id=job.operation_id
              AND authority.status='completed'
       );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='protected authority delivery lease is stale';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_delivery_retry_v1(
    uuid,uuid,bigint,text,boolean) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_delivery_await_import_v1(
    uuid,uuid,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_delivery_succeed_v1(
    uuid,uuid,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_retry_v1(
    uuid,uuid,bigint,text,boolean) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_await_import_v1(
    uuid,uuid,bigint) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_succeed_v1(
    uuid,uuid,bigint) TO forgecrud_runtime;
