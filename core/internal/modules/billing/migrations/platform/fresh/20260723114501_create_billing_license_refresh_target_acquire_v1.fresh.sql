-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_refresh_target_acquire_v1()
RETURNS TABLE (
    target_found boolean, ready boolean, operation_id uuid,
    license_id uuid, licensee_organization_id uuid, environment_id uuid,
    delivery_kind text, target_revision bigint,
    retry_at timestamptz, database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_head platform.billing_license_heads%ROWTYPE;
    v_snapshot platform.billing_license_snapshots%ROWTYPE;
    v_target platform.billing_license_refresh_targets%ROWTYPE;
    v_operation_state text;
    v_count integer;
    v_ready boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR pg_catalog.current_setting('application_name', true)
            <> 'forgecrud-v2:core-worker' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing license refresh target denied';
    END IF;
    IF NOT EXISTS (
        SELECT 1
          FROM platform.tenancy_deployment_policy AS policy
          JOIN platform.tenancy_tenant_bindings AS binding
            ON binding.environment_id = policy.environment_id
           AND binding.purpose = 'onpremise_primary'
         WHERE policy.singleton
           AND policy.deployment_mode = 'onpremise'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing license refresh requires onpremise primary';
    END IF;
    SELECT pg_catalog.count(*) INTO v_count
      FROM platform.billing_license_heads AS head
      JOIN platform.billing_license_snapshots AS snapshot
        ON snapshot.snapshot_id = head.snapshot_id
      JOIN platform.tenancy_deployment_policy AS policy
        ON policy.environment_id = head.environment_id
       AND policy.singleton
       AND policy.deployment_mode = 'onpremise'
      JOIN platform.tenancy_tenant_bindings AS binding
        ON binding.environment_id = policy.environment_id
       AND binding.purpose = 'onpremise_primary';
    IF v_count = 0 THEN
        RETURN QUERY SELECT false, false,
            '00000000-0000-0000-0000-000000000000'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid,
            ''::text, 0::bigint, v_now + interval '30 seconds', v_now;
        RETURN;
    END IF;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh target is ambiguous';
    END IF;
    SELECT head.* INTO v_head
      FROM platform.billing_license_heads AS head
      JOIN platform.tenancy_deployment_policy AS policy
        ON policy.environment_id = head.environment_id
       AND policy.singleton
       AND policy.deployment_mode = 'onpremise'
      JOIN platform.tenancy_tenant_bindings AS binding
        ON binding.environment_id = policy.environment_id
       AND binding.purpose = 'onpremise_primary'
     FOR SHARE OF head;
    SELECT snapshot.* INTO v_snapshot
      FROM platform.billing_license_snapshots AS snapshot
     WHERE snapshot.snapshot_id = v_head.snapshot_id
     FOR SHARE;
    INSERT INTO platform.billing_license_refresh_targets (
        license_id, environment_id, licensee_organization_id, delivery_kind,
        observed_sequence, observed_sha256, operation_id,
        next_attempt_at, lease_until, revision, updated_at
    ) VALUES (
        v_head.license_id, v_head.environment_id,
        v_snapshot.licensee_organization_id, v_snapshot.delivery_kind,
        v_head.sequence, v_head.payload_sha256, pg_catalog.uuidv7(),
        v_snapshot.refresh_after, NULL, 1, v_now
    ) ON CONFLICT ON CONSTRAINT billing_license_refresh_targets_pkey
      DO NOTHING;
    SELECT * INTO v_target
      FROM platform.billing_license_refresh_targets AS target
     WHERE target.license_id = v_head.license_id
       AND target.environment_id = v_head.environment_id
     FOR UPDATE;
    IF v_target.observed_sequence IS DISTINCT FROM v_head.sequence
       OR v_target.observed_sha256 IS DISTINCT FROM v_head.payload_sha256 THEN
        UPDATE platform.billing_license_refresh_targets AS refresh_target
           SET licensee_organization_id = v_snapshot.licensee_organization_id,
               delivery_kind = v_snapshot.delivery_kind,
               observed_sequence = v_head.sequence,
               observed_sha256 = v_head.payload_sha256,
               operation_id = pg_catalog.uuidv7(),
               next_attempt_at = v_snapshot.refresh_after,
               lease_until = NULL,
               revision = refresh_target.revision + 1, updated_at = v_now
         WHERE refresh_target.license_id = v_head.license_id
           AND refresh_target.environment_id = v_head.environment_id
         RETURNING refresh_target.* INTO v_target;
    END IF;
    SELECT operation.state INTO v_operation_state
      FROM platform.billing_license_refresh_operations AS operation
     WHERE operation.operation_id = v_target.operation_id;
    IF v_operation_state = 'fenced' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh target is fenced';
    END IF;
    IF v_operation_state = 'succeeded'
       AND v_target.next_attempt_at <= v_now THEN
        UPDATE platform.billing_license_refresh_targets AS refresh_target
           SET operation_id = pg_catalog.uuidv7(), lease_until = NULL,
               revision = refresh_target.revision + 1, updated_at = v_now
         WHERE refresh_target.license_id = v_target.license_id
           AND refresh_target.environment_id = v_target.environment_id
         RETURNING refresh_target.* INTO v_target;
    END IF;
    v_ready := v_target.next_attempt_at <= v_now
        AND (v_target.lease_until IS NULL OR v_target.lease_until <= v_now);
    IF v_ready THEN
        UPDATE platform.billing_license_refresh_targets AS refresh_target
           SET lease_until = v_now + interval '60 seconds',
               next_attempt_at = v_now + interval '5 minutes',
               revision = refresh_target.revision + 1, updated_at = v_now
         WHERE refresh_target.license_id = v_target.license_id
           AND refresh_target.environment_id = v_target.environment_id
         RETURNING refresh_target.* INTO v_target;
    END IF;
    RETURN QUERY SELECT true, v_ready, v_target.operation_id,
        v_target.license_id, v_target.licensee_organization_id,
        v_target.environment_id, v_target.delivery_kind, v_target.revision,
        CASE WHEN v_ready THEN v_target.lease_until
             ELSE GREATEST(v_target.next_attempt_at,
                 COALESCE(v_target.lease_until, v_target.next_attempt_at))
        END, v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_refresh_target_acquire_v1()
FROM PUBLIC;
