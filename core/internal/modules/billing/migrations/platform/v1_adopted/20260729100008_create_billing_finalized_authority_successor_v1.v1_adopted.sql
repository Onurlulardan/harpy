-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_finalize_authority_successor_v1(
    p_operation_id uuid, p_run_id uuid, p_forms_operation_id uuid,
    p_forms_run_id uuid, p_acceptance_digest text
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_companion platform.cutover_operations%ROWTYPE;
    v_group_generation bigint;
    v_requested_by uuid;
    v_approved_by uuid;
    v_reason text;
    v_change_reference text;
    v_changed_at timestamptz;
    v_members jsonb;
    v_authority_id uuid;
    v_locked_count integer;
    v_run_hash text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_operation_id IS NULL OR p_run_id IS NULL
       OR pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_run_id) IS DISTINCT FROM 7
       OR p_acceptance_digest !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing finalized authority successor is denied';
    END IF;
    SELECT operation.* INTO STRICT v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id
      AND operation.run_id = p_run_id
      AND operation.state = 'completed'
      AND operation.domain = 'billing'
      AND (
          (operation.scope_kind = 'global' AND operation.group_id = 'billing')
          OR (
              operation.scope_kind = 'tenant'
              AND operation.group_id = 'billing_forms'
          )
      )
    FOR SHARE;
    IF v_operation.scope_kind = 'global' THEN
        IF p_forms_operation_id IS NOT NULL OR p_forms_run_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing global successor companion is invalid';
        END IF;
    ELSE
        IF p_forms_operation_id IS NULL OR p_forms_run_id IS NULL
           OR pg_catalog.uuid_extract_version(p_forms_operation_id) IS DISTINCT FROM 7
           OR pg_catalog.uuid_extract_version(p_forms_run_id) IS DISTINCT FROM 7 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing tenant successor companion is required';
        END IF;
        SELECT operation.* INTO STRICT v_companion
        FROM platform.cutover_operations AS operation
        WHERE operation.operation_id = p_forms_operation_id
          AND operation.run_id = p_forms_run_id
          AND operation.state = 'completed'
          AND operation.scope_kind = v_operation.scope_kind
          AND operation.scope_id = v_operation.scope_id
          AND operation.group_id = v_operation.group_id
          AND operation.domain = 'forms'
          AND operation.cutover_generation = v_operation.cutover_generation
        FOR SHARE;
    END IF;
    SELECT group_row.generation, group_row.requested_by,
           group_row.approved_by, group_row.reason,
           group_row.change_reference, group_row.changed_at
    INTO STRICT v_group_generation, v_requested_by, v_approved_by,
        v_reason, v_change_reference, v_changed_at
    FROM platform.cutover_groups AS group_row
    WHERE group_row.scope_kind = v_operation.scope_kind
      AND group_row.scope_id = v_operation.scope_id
      AND group_row.group_id = v_operation.group_id
      AND group_row.state = 'v2' AND group_row.writer = 'v2'
      AND group_row.generation = v_operation.cutover_generation + 1
      AND group_row.acceptance_digest = p_acceptance_digest
      AND group_row.approved_by IS NOT NULL
      AND group_row.approved_by <> group_row.requested_by
    FOR SHARE;
    SELECT pg_catalog.jsonb_agg(member.domain ORDER BY member.ordinal)
    INTO v_members
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = v_operation.scope_kind
      AND member.scope_id = v_operation.scope_id
      AND member.group_id = v_operation.group_id;
    IF (
        v_operation.scope_kind = 'global'
        AND v_members IS DISTINCT FROM '["billing"]'::jsonb
    ) OR (
        v_operation.scope_kind = 'tenant'
        AND v_members IS DISTINCT FROM '["billing","forms"]'::jsonb
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing finalized authority members differ';
    END IF;
    PERFORM 1
    FROM platform.cutover_group_members AS member
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member.scope_kind
     AND domain_row.scope_id = member.scope_id
     AND domain_row.domain = member.domain
    WHERE member.scope_kind = v_operation.scope_kind
      AND member.scope_id = v_operation.scope_id
      AND member.group_id = v_operation.group_id
      AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
      AND domain_row.generation = v_group_generation
      AND domain_row.acceptance_digest = p_acceptance_digest
      AND domain_row.requested_by = v_requested_by
      AND domain_row.approved_by = v_approved_by
      AND domain_row.reason = v_reason
      AND domain_row.change_reference = v_change_reference
      AND domain_row.changed_at = v_changed_at
    ORDER BY member.ordinal
    FOR SHARE OF member, domain_row;
    GET DIAGNOSTICS v_locked_count = ROW_COUNT;
    IF (v_operation.scope_kind = 'global' AND v_locked_count IS DISTINCT FROM 1)
       OR (v_operation.scope_kind = 'tenant' AND v_locked_count IS DISTINCT FROM 2)
    THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing finalized authority domains differ';
    END IF;
    v_run_hash := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(p_run_id::text, 'UTF8')), 'hex');
    v_authority_id := (pg_catalog.substring(p_operation_id::text, 1, 19) ||
        '8' || pg_catalog.substring(v_run_hash, 1, 3) || '-' ||
        pg_catalog.substring(v_run_hash, 4, 12))::uuid;
    IF v_operation.scope_kind = 'global' THEN
        PERFORM 1
        FROM platform.billing_global_authorities AS authority
        WHERE authority.authority_id = v_authority_id
          AND authority.scope_id = v_operation.scope_id
          AND authority.authority_group_generation = v_group_generation
          AND authority.authority_domain_generation = v_group_generation;
    ELSE
        PERFORM 1
        FROM platform.billing_tenant_authorities AS authority
        WHERE authority.authority_id = v_authority_id
          AND authority.tenant_id = v_operation.scope_id
          AND authority.authority_group_generation = v_group_generation
          AND authority.billing_generation = v_group_generation
          AND authority.forms_generation = v_group_generation;
    END IF;
    IF FOUND THEN
        RETURN v_authority_id;
    END IF;
    BEGIN
        IF v_operation.scope_kind = 'global' THEN
            INSERT INTO platform.billing_global_authorities (
                authority_id, scope_kind, scope_id, authority_group_id,
                authority_group_generation, authority_domain,
                authority_domain_generation, created_at
            ) VALUES (
                v_authority_id, 'global', v_operation.scope_id, 'billing',
                v_group_generation, 'billing', v_group_generation,
                pg_catalog.transaction_timestamp()
            );
        ELSE
            INSERT INTO platform.billing_tenant_authorities (
                authority_id, tenant_id, scope_kind, scope_id,
                authority_group_id, authority_group_generation,
                billing_domain, billing_generation,
                forms_domain, forms_generation, created_at
            ) VALUES (
                v_authority_id, v_operation.scope_id, 'tenant',
                v_operation.scope_id, 'billing_forms', v_group_generation,
                'billing', v_group_generation, 'forms', v_group_generation,
                pg_catalog.transaction_timestamp()
            );
        END IF;
    EXCEPTION WHEN SQLSTATE '23505' THEN
        IF v_operation.scope_kind = 'global' THEN
            SELECT authority.authority_id INTO STRICT v_authority_id
            FROM platform.billing_global_authorities AS authority
            WHERE authority.authority_id = v_authority_id
              AND authority.scope_id = v_operation.scope_id
              AND authority.authority_group_generation = v_group_generation
              AND authority.authority_domain_generation = v_group_generation;
        ELSE
            SELECT authority.authority_id INTO STRICT v_authority_id
            FROM platform.billing_tenant_authorities AS authority
            WHERE authority.authority_id = v_authority_id
              AND authority.tenant_id = v_operation.scope_id
              AND authority.authority_group_generation = v_group_generation
              AND authority.billing_generation = v_group_generation
              AND authority.forms_generation = v_group_generation;
        END IF;
    END;
    RETURN v_authority_id;
EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
    RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing finalized authority evidence differs';
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.billing_finalize_authority_successor_v1(
    uuid,uuid,uuid,uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_finalize_authority_successor_v1(
    uuid,uuid,uuid,uuid,text) TO harpy_cutover_authority_owner;
