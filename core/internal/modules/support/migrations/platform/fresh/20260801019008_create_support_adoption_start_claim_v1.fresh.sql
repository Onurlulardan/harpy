-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_start_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_source_kind text,
    p_source_tenant_id uuid,
    p_source_schema text,
    p_source_relation text,
    p_source_relation_oid oid,
    p_tenant_slug text,
    p_tenant_name text,
    p_source_locator_sha256 bytea,
    p_source_manifest_sha256 bytea,
    p_transform_sha256 bytea,
    p_snapshot_id text,
    p_expected_count bigint,
    p_expected_source_sha256 bytea,
    p_observed_at timestamp(6) with time zone
)
RETURNS TABLE(run_document jsonb, replayed boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
    v_run platform.support_adoption_runs%ROWTYPE;
BEGIN
    PERFORM platform.support_require_adopter_v1();
    IF p_operation_id IS NULL OR p_operation_id = '00000000-0000-0000-0000-000000000000'
       OR p_run_id IS NULL OR pg_catalog.uuid_extract_version(p_run_id) <> 7
       OR p_expected_count < 0 OR p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at
       OR pg_catalog.octet_length(p_source_locator_sha256) <> 32
       OR pg_catalog.octet_length(p_source_manifest_sha256) <> 32
       OR pg_catalog.octet_length(p_transform_sha256) <> 32
       OR pg_catalog.octet_length(p_expected_source_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support adoption plan';
    END IF;
    SELECT run.* INTO v_run FROM platform.support_adoption_runs AS run
    WHERE run.operation_id=p_operation_id;
    IF FOUND THEN
        IF v_run.run_id <> p_run_id OR v_run.source_kind <> p_source_kind
           OR v_run.source_tenant_id IS DISTINCT FROM p_source_tenant_id
           OR v_run.source_schema <> p_source_schema
           OR v_run.source_relation <> p_source_relation
           OR v_run.source_relation_oid <> p_source_relation_oid
           OR v_run.tenant_slug <> p_tenant_slug OR v_run.tenant_name <> p_tenant_name
           OR v_run.source_locator_sha256 <> p_source_locator_sha256
           OR v_run.source_manifest_sha256 <> p_source_manifest_sha256
           OR v_run.transform_sha256 <> p_transform_sha256
           OR v_run.snapshot_id <> p_snapshot_id
           OR v_run.expected_count <> p_expected_count
           OR v_run.expected_source_sha256 <> p_expected_source_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption plan drift';
        END IF;
        RETURN QUERY SELECT platform.support_adoption_run_document_v1(v_run), true;
        RETURN;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
        WHERE relation.oid=p_source_relation_oid
          AND namespace.nspname=p_source_schema
          AND relation.relname=p_source_relation
          AND relation.relkind IN ('r','p')
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support source locator drift';
    END IF;

    INSERT INTO platform.support_adoption_runs (
        run_id,operation_id,source_kind,source_tenant_id,source_schema,source_relation,
        source_relation_oid,tenant_slug,tenant_name,source_locator_sha256,
        source_manifest_sha256,transform_sha256,snapshot_id,expected_count,
        expected_source_sha256,copied_count,source_sha256,target_sha256,status,
        revision,next_page,lease_generation,created_at,updated_at
    ) VALUES (
        p_run_id,p_operation_id,p_source_kind,p_source_tenant_id,p_source_schema,
        p_source_relation,p_source_relation_oid,p_tenant_slug,p_tenant_name,
        p_source_locator_sha256,p_source_manifest_sha256,p_transform_sha256,
        p_snapshot_id,p_expected_count,p_expected_source_sha256,0,
        pg_catalog.sha256(''::bytea),pg_catalog.sha256(''::bytea),'planned',
        1,1,0,p_observed_at,p_observed_at
    ) RETURNING * INTO v_run;
    RETURN QUERY SELECT platform.support_adoption_run_document_v1(v_run), false;
END;
$$;

CREATE FUNCTION platform.support_claim_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_lease_owner uuid,
    p_expected_revision bigint,
    p_observed_at timestamp(6) with time zone
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
    v_run platform.support_adoption_runs%ROWTYPE;
BEGIN
    PERFORM platform.support_require_adopter_v1();
    IF p_lease_owner IS NULL OR p_lease_owner='00000000-0000-0000-0000-000000000000'
       OR p_expected_revision < 1 OR p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support adoption claim';
    END IF;
    SELECT run.* INTO v_run FROM platform.support_adoption_runs AS run
    WHERE run.operation_id=p_operation_id AND run.run_id=p_run_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6514', MESSAGE='support adoption run not found';
    END IF;
    IF v_run.status='completed' THEN
        IF v_run.lease_owner <> p_lease_owner THEN
            RAISE EXCEPTION USING ERRCODE='P6513', MESSAGE='support adoption lease mismatch';
        END IF;
        RETURN platform.support_adoption_run_document_v1(v_run);
    END IF;
    IF v_run.status='fenced' THEN
        RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption run is fenced';
    END IF;
    IF v_run.revision = p_expected_revision + 1
       AND v_run.lease_claim_revision = p_expected_revision
       AND v_run.lease_owner = p_lease_owner AND v_run.lease_expires_at > p_observed_at THEN
        RETURN platform.support_adoption_run_document_v1(v_run);
    END IF;
    IF v_run.revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE='P6511', MESSAGE='support adoption revision mismatch';
    END IF;
    IF v_run.lease_owner IS NOT NULL AND v_run.lease_owner <> p_lease_owner
       AND v_run.lease_expires_at > p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6513', MESSAGE='support adoption lease held';
    END IF;
    UPDATE platform.support_adoption_runs AS run SET
        status=CASE WHEN run.status='planned' THEN 'applying' ELSE run.status END,
        lease_owner=p_lease_owner, lease_claim_revision=p_expected_revision,
        lease_generation=run.lease_generation+1,
        lease_expires_at=p_observed_at+interval '5 minutes',
        revision=run.revision+1, updated_at=p_observed_at
    WHERE run.run_id=p_run_id RETURNING * INTO v_run;
    RETURN platform.support_adoption_run_document_v1(v_run);
END;
$$;

REVOKE ALL ON FUNCTION platform.support_start_adoption_v1(uuid,uuid,text,uuid,text,text,oid,text,text,bytea,bytea,bytea,text,bigint,bytea,timestamp with time zone) FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON FUNCTION platform.support_claim_adoption_v1(uuid,uuid,uuid,bigint,timestamp with time zone) FROM PUBLIC, forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_start_adoption_v1(uuid,uuid,text,uuid,text,text,oid,text,text,bytea,bytea,bytea,text,bigint,bytea,timestamp with time zone) TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION platform.support_claim_adoption_v1(uuid,uuid,uuid,bigint,timestamp with time zone) TO forgecrud_adopter;
-- +goose StatementEnd
