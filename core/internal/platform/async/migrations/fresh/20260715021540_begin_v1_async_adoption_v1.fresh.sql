-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_begin_v1_adoption_v1(
    p_run_id uuid,
    p_scope_kind text,
    p_scope_id uuid,
    p_domain text,
    p_group_id text,
    p_source_generation bigint,
    p_target_generation bigint,
    p_source_group_generation bigint,
    p_target_group_generation bigint,
    p_inventory_sha256 text,
    p_component_manifest_sha256 text,
    p_expected_enqueue_fences smallint,
    p_expected_scheduler_fences smallint,
    p_expected_consumer_fences smallint,
    p_item_count bigint,
    p_captured_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group_state text;
    v_group_writer text;
    v_group_generation bigint;
    v_domain_state text;
    v_domain_writer text;
    v_domain_generation bigint;
    v_inserted boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5632',
            MESSAGE = 'V1 async adoption requires adopter authority';
    END IF;
    SELECT state, writer, generation
    INTO v_group_state, v_group_writer, v_group_generation
    FROM platform.cutover_groups
    WHERE scope_kind OPERATOR(pg_catalog.=) p_scope_kind
      AND scope_id OPERATOR(pg_catalog.=) p_scope_id
      AND group_id OPERATOR(pg_catalog.=) p_group_id
    FOR SHARE;
    IF NOT FOUND OR v_group_state IS DISTINCT FROM 'quiescing' OR
       v_group_writer IS DISTINCT FROM 'none' OR
       v_group_generation IS DISTINCT FROM p_source_group_generation OR
       p_target_group_generation IS DISTINCT FROM p_source_group_generation + 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5633',
            MESSAGE = 'V1 async adoption group is not quiesced';
    END IF;
    SELECT state, writer, generation
    INTO v_domain_state, v_domain_writer, v_domain_generation
    FROM platform.domain_cutovers
    WHERE scope_kind OPERATOR(pg_catalog.=) p_scope_kind
      AND scope_id OPERATOR(pg_catalog.=) p_scope_id
      AND domain OPERATOR(pg_catalog.=) p_domain
    FOR SHARE;
    IF NOT FOUND OR v_domain_state IS DISTINCT FROM 'quiescing' OR
       v_domain_writer IS DISTINCT FROM 'none' OR
       v_domain_generation IS DISTINCT FROM p_source_generation OR
       p_target_generation IS DISTINCT FROM p_source_generation + 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5634',
            MESSAGE = 'V1 async adoption domain is not quiesced';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members
        WHERE scope_kind OPERATOR(pg_catalog.=) p_scope_kind
          AND scope_id OPERATOR(pg_catalog.=) p_scope_id
          AND group_id OPERATOR(pg_catalog.=) p_group_id
          AND domain OPERATOR(pg_catalog.=) p_domain
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5635',
            MESSAGE = 'V1 async adoption group membership is missing';
    END IF;
    INSERT INTO platform.async_v1_adoption_runs (
        run_id, scope_kind, scope_id, domain, group_id,
        source_generation, target_generation,
        source_group_generation, target_group_generation,
        inventory_contract_version, inventory_revision, inventory_sha256,
        component_manifest_sha256, expected_enqueue_fences,
        expected_scheduler_fences, expected_consumer_fences,
        item_count, state, captured_at
    ) VALUES (
        p_run_id, p_scope_kind, p_scope_id, p_domain, p_group_id,
        p_source_generation, p_target_generation,
        p_source_group_generation, p_target_group_generation,
        'async-adoption-inventory.v1', 1, p_inventory_sha256,
        p_component_manifest_sha256, p_expected_enqueue_fences,
        p_expected_scheduler_fences, p_expected_consumer_fences,
        p_item_count, 'inventory', p_captured_at
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM platform.async_v1_adoption_runs AS run
        WHERE run.run_id OPERATOR(pg_catalog.=) p_run_id
          AND run.scope_kind OPERATOR(pg_catalog.=) p_scope_kind
          AND run.scope_id OPERATOR(pg_catalog.=) p_scope_id
          AND run.domain OPERATOR(pg_catalog.=) p_domain
          AND run.group_id OPERATOR(pg_catalog.=) p_group_id
          AND run.source_generation OPERATOR(pg_catalog.=) p_source_generation
          AND run.target_generation OPERATOR(pg_catalog.=) p_target_generation
          AND run.source_group_generation OPERATOR(pg_catalog.=) p_source_group_generation
          AND run.target_group_generation OPERATOR(pg_catalog.=) p_target_group_generation
          AND run.inventory_sha256 OPERATOR(pg_catalog.=) p_inventory_sha256
          AND run.component_manifest_sha256 OPERATOR(pg_catalog.=) p_component_manifest_sha256
          AND run.item_count OPERATOR(pg_catalog.=) p_item_count
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5636',
            MESSAGE = 'V1 async adoption run replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_begin_v1_adoption_v1(
    uuid,text,uuid,text,text,bigint,bigint,bigint,bigint,text,text,
    smallint,smallint,smallint,bigint,timestamptz
) FROM PUBLIC;
