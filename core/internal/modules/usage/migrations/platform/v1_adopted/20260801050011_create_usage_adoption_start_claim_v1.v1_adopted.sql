-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_adoption_run_document_v1(p_run_id uuid)
RETURNS jsonb LANGUAGE sql SECURITY DEFINER SET search_path TO pg_catalog AS $$
SELECT pg_catalog.jsonb_build_object(
    'run_id', runs.run_id, 'operation_id', runs.operation_id,
    'source_kind', runs.source_kind, 'source_schema', runs.source_schema,
    'source_relation', runs.source_relation,
    'source_relation_oid', runs.source_relation_oid::bigint,
    'source_count', runs.source_count,
    'source_root_sha256', pg_catalog.encode(runs.source_root_sha256, 'hex'),
    'snapshot_id', runs.snapshot_id,
    'transform_sha256', pg_catalog.encode(runs.transform_sha256, 'hex'),
    'next_page', runs.next_page, 'cursor_id', runs.cursor_id,
    'copied_count', runs.copied_count, 'status', runs.status,
    'lease_owner', runs.lease_owner, 'lease_generation', runs.lease_generation,
    'lease_expires_at', runs.lease_expires_at, 'revision', runs.revision,
    'completed_at', runs.completed_at)
FROM platform.usage_adoption_runs AS runs WHERE runs.run_id = p_run_id
$$;

CREATE FUNCTION platform.usage_adoption_start_v1(
    p_operation_id uuid, p_run_id uuid, p_source_kind text, p_source_schema text,
    p_source_relation text, p_source_relation_oid oid, p_source_count bigint,
    p_source_root_sha256 bytea, p_snapshot_id text, p_transform_sha256 bytea,
    p_observed_at timestamptz
) RETURNS TABLE(document jsonb, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_existing platform.usage_adoption_runs%ROWTYPE;
BEGIN
    IF p_operation_id IS NULL OR p_run_id IS NULL
       OR pg_catalog.uuid_extract_version(p_run_id) <> 7
       OR p_source_kind NOT IN ('zero','legacy_trace_logs')
       OR p_source_schema !~ '^[a-z][a-z0-9_]{0,62}$'
       OR p_source_relation !~ '^[a-z][a-z0-9_]{0,62}$'
       OR p_source_count < 0 OR pg_catalog.octet_length(p_source_root_sha256) <> 32
       OR pg_catalog.octet_length(p_transform_sha256) <> 32
       OR pg_catalog.octet_length(p_snapshot_id) NOT BETWEEN 16 AND 200
       OR p_observed_at IS NULL
       OR (p_source_kind = 'zero' AND (p_source_relation_oid <> 0 OR p_source_count <> 0))
       OR (p_source_kind = 'legacy_trace_logs' AND p_source_relation_oid::bigint < 1) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage adoption start is invalid';
    END IF;
    INSERT INTO platform.usage_adoption_runs AS target (
        run_id, operation_id, source_kind, source_schema, source_relation,
        source_relation_oid, source_count, source_root_sha256, snapshot_id,
        transform_sha256, status, created_at, updated_at
    ) VALUES (p_run_id, p_operation_id, p_source_kind, p_source_schema,
        p_source_relation, p_source_relation_oid, p_source_count, p_source_root_sha256,
        p_snapshot_id, p_transform_sha256, 'planned', p_observed_at, p_observed_at)
    ON CONFLICT (operation_id) DO NOTHING;
    SELECT * INTO v_existing FROM platform.usage_adoption_runs AS runs
    WHERE runs.operation_id = p_operation_id;
    IF v_existing.run_id IS DISTINCT FROM p_run_id
       OR v_existing.source_kind IS DISTINCT FROM p_source_kind
       OR v_existing.source_schema IS DISTINCT FROM p_source_schema
       OR v_existing.source_relation IS DISTINCT FROM p_source_relation
       OR v_existing.source_relation_oid IS DISTINCT FROM p_source_relation_oid
       OR v_existing.source_count IS DISTINCT FROM p_source_count
       OR v_existing.source_root_sha256 IS DISTINCT FROM p_source_root_sha256
       OR v_existing.snapshot_id IS DISTINCT FROM p_snapshot_id
       OR v_existing.transform_sha256 IS DISTINCT FROM p_transform_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption start conflicts';
    END IF;
    document := platform.usage_adoption_run_document_v1(v_existing.run_id);
    replayed := v_existing.created_at IS DISTINCT FROM p_observed_at;
    RETURN NEXT;
END;
$$;

CREATE FUNCTION platform.usage_adoption_claim_v1(
    p_operation_id uuid, p_run_id uuid, p_lease_owner uuid, p_expected_revision bigint,
    p_source_relation_oid oid, p_source_count bigint, p_source_root_sha256 bytea,
    p_observed_at timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_run platform.usage_adoption_runs%ROWTYPE;
BEGIN
    SELECT * INTO v_run FROM platform.usage_adoption_runs AS runs
    WHERE runs.operation_id = p_operation_id AND runs.run_id = p_run_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage adoption missing'; END IF;
    IF v_run.source_relation_oid IS DISTINCT FROM p_source_relation_oid
       OR v_run.source_count IS DISTINCT FROM p_source_count
       OR v_run.source_root_sha256 IS DISTINCT FROM p_source_root_sha256 THEN
        UPDATE platform.usage_adoption_runs SET status = 'fenced', lease_owner = NULL,
            lease_expires_at = NULL, revision = revision + 1, updated_at = p_observed_at
        WHERE run_id = p_run_id;
        RETURN platform.usage_adoption_run_document_v1(p_run_id);
    END IF;
    IF v_run.status IN ('completed','fenced') THEN
        RETURN platform.usage_adoption_run_document_v1(p_run_id);
    END IF;
    IF p_lease_owner IS NULL OR p_expected_revision <> v_run.revision OR p_observed_at IS NULL
       OR (v_run.lease_owner IS NOT NULL AND v_run.lease_owner <> p_lease_owner
           AND v_run.lease_expires_at > p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption lease conflicts';
    END IF;
    UPDATE platform.usage_adoption_runs SET status = 'applying', lease_owner = p_lease_owner,
        lease_generation = lease_generation + 1, lease_expires_at = p_observed_at + interval '30 seconds',
        revision = revision + 1, updated_at = p_observed_at WHERE run_id = p_run_id;
    RETURN platform.usage_adoption_run_document_v1(p_run_id);
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_adoption_run_document_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_adoption_start_v1(
uuid,uuid,text,text,text,oid,bigint,bytea,text,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_adoption_claim_v1(
uuid,uuid,uuid,bigint,oid,bigint,bytea,timestamptz) FROM PUBLIC;
-- +goose StatementEnd
