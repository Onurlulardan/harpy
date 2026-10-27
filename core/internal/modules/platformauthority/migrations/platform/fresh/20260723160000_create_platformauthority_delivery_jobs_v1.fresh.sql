-- +goose Up
CREATE TABLE platform.platformauthority_delivery_jobs (
    operation_id uuid PRIMARY KEY
        REFERENCES platform.platformauthority_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    state text NOT NULL CHECK (
        state IN ('queued','running','awaiting_import','succeeded','dead')
    ),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts BETWEEN 0 AND 8),
    max_attempts integer NOT NULL DEFAULT 8 CHECK (max_attempts = 8),
    lease_generation bigint NOT NULL DEFAULT 0 CHECK (lease_generation >= 0),
    lease_owner_id uuid,
    lease_until timestamptz,
    next_attempt_at timestamptz NOT NULL,
    last_error_code text CHECK (
        last_error_code ~ '^[A-Z][A-Z0-9_]{0,127}$'
    ),
    completed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK (
        (state = 'running' AND lease_owner_id IS NOT NULL AND lease_until IS NOT NULL)
        OR (state <> 'running' AND lease_owner_id IS NULL AND lease_until IS NULL)
    ),
    CHECK (
        (state = 'succeeded' AND completed_at IS NOT NULL)
        OR (state <> 'succeeded' AND completed_at IS NULL)
    )
);

CREATE INDEX platformauthority_delivery_due_idx
ON platform.platformauthority_delivery_jobs (next_attempt_at, operation_id)
WHERE state IN ('queued','awaiting_import');

-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_delivery_claim_v1(p_lease_owner_id uuid)
RETURNS TABLE(operation_id uuid,lease_generation bigint,attempt integer,max_attempts integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_operation_id uuid;
BEGIN
    IF p_lease_owner_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='protected authority delivery owner is invalid';
    END IF;
    INSERT INTO platform.platformauthority_delivery_jobs (
        operation_id,state,next_attempt_at,created_at,updated_at
    )
    SELECT authority.operation_id,'queued',v_now,v_now,v_now
      FROM platform.platformauthority_operations AS authority
     WHERE authority.purpose NOT IN (
        'initial_environment_authority','initial_tenant_authority'
     )
    ON CONFLICT ON CONSTRAINT platformauthority_delivery_jobs_pkey DO NOTHING;

    UPDATE platform.platformauthority_delivery_jobs AS job
       SET state=CASE WHEN job.attempts>=job.max_attempts THEN 'dead' ELSE 'queued' END,
           lease_owner_id=NULL,lease_until=NULL,
           next_attempt_at=CASE WHEN job.attempts>=job.max_attempts
               THEN job.next_attempt_at ELSE v_now END,
           last_error_code='LEASE_EXPIRED',updated_at=v_now
     WHERE job.state='running' AND job.lease_until<=v_now;

    SELECT job.operation_id INTO v_operation_id
      FROM platform.platformauthority_delivery_jobs AS job
      JOIN platform.platformauthority_operations AS authority
        ON authority.operation_id=job.operation_id
     WHERE (job.state='queued' AND job.next_attempt_at<=v_now)
        OR (job.state='awaiting_import' AND authority.status='completed')
     ORDER BY job.next_attempt_at,job.operation_id
     FOR UPDATE OF job SKIP LOCKED
     LIMIT 1;
    IF v_operation_id IS NULL THEN RETURN; END IF;

    RETURN QUERY
    UPDATE platform.platformauthority_delivery_jobs AS job
       SET state='running',attempts=job.attempts+1,
           lease_generation=job.lease_generation+1,
           lease_owner_id=p_lease_owner_id,lease_until=v_now+interval '60 seconds',
           last_error_code=NULL,updated_at=v_now
     WHERE job.operation_id=v_operation_id
     RETURNING job.operation_id,job.lease_generation,job.attempts,job.max_attempts;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_delivery_claim_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_claim_v1(uuid)
TO harpy_runtime;
