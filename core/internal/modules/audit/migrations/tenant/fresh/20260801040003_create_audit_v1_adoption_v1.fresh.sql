-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION audit_v1_derived_id_v1(p_source_id uuid, p_label text) RETURNS uuid
LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog AS $$
DECLARE
    v_digest bytea;
    v_hex text;
BEGIN
    IF p_source_id IS NULL OR p_source_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_label !~ '^[a-z0-9_-]{1,64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'AUDIT_V1_IDENTITY_INVALID';
    END IF;
    v_digest := pg_catalog.sha256(
        pg_catalog.convert_to('harpy:audit:v1-adoption:v1','UTF8') ||
        pg_catalog.uuid_send(p_source_id) || '\x00'::bytea ||
        pg_catalog.convert_to(p_label,'UTF8'));
    v_digest := pg_catalog.set_byte(v_digest, 6,
        (pg_catalog.get_byte(v_digest, 6) & 15) | 112);
    v_digest := pg_catalog.set_byte(v_digest, 8,
        (pg_catalog.get_byte(v_digest, 8) & 63) | 128);
    v_hex := pg_catalog.encode(pg_catalog.substring(v_digest, 1, 16), 'hex');
    RETURN (pg_catalog.substring(v_hex,1,8) || '-' || pg_catalog.substring(v_hex,9,4) ||
        '-' || pg_catalog.substring(v_hex,13,4) || '-' ||
        pg_catalog.substring(v_hex,17,4) || '-' || pg_catalog.substring(v_hex,21,12))::uuid;
END;
$$;

CREATE FUNCTION audit_v1_redacted_changes_v1(
    p_old jsonb, p_new jsonb, p_changed jsonb
) RETURNS jsonb
LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog AS $$
DECLARE
    v_old jsonb := CASE WHEN pg_catalog.jsonb_typeof(p_old)='object' THEN p_old ELSE '{}'::jsonb END;
    v_new jsonb := CASE WHEN pg_catalog.jsonb_typeof(p_new)='object' THEN p_new ELSE '{}'::jsonb END;
    v_changed jsonb := CASE WHEN pg_catalog.jsonb_typeof(p_changed)='array' THEN p_changed ELSE '[]'::jsonb END;
    v_result jsonb := '[]'::jsonb;
    v_seen jsonb := '{}'::jsonb;
    v_field text;
    v_safe text;
    v_item jsonb;
    v_before bytea;
    v_after bytea;
BEGIN
    FOR v_field IN
        SELECT key FROM (
            SELECT key FROM pg_catalog.jsonb_object_keys(v_old) AS old_keys(key)
            UNION SELECT key FROM pg_catalog.jsonb_object_keys(v_new) AS new_keys(key)
            UNION SELECT value->>'field' FROM pg_catalog.jsonb_array_elements(v_changed)
                  WHERE pg_catalog.jsonb_typeof(value)='object' AND value ? 'field'
        ) AS fields WHERE key IS NOT NULL ORDER BY key LIMIT 100
    LOOP
        SELECT value INTO v_item FROM pg_catalog.jsonb_array_elements(v_changed)
         WHERE value->>'field'=v_field LIMIT 1;
        v_before := NULL;
        v_after := NULL;
        IF v_item IS NOT NULL THEN
            IF v_item ? 'old_value' THEN
                v_before := pg_catalog.sha256(pg_catalog.convert_to((v_item->'old_value')::text,'UTF8'));
            END IF;
            IF v_item ? 'new_value' THEN
                v_after := pg_catalog.sha256(pg_catalog.convert_to((v_item->'new_value')::text,'UTF8'));
            END IF;
        ELSE
            IF v_old ? v_field THEN
                v_before := pg_catalog.sha256(pg_catalog.convert_to((v_old->v_field)::text,'UTF8'));
            END IF;
            IF v_new ? v_field THEN
                v_after := pg_catalog.sha256(pg_catalog.convert_to((v_new->v_field)::text,'UTF8'));
            END IF;
        END IF;
        IF v_before IS NULL AND v_after IS NULL OR v_before IS NOT NULL AND v_before=v_after THEN
            CONTINUE;
        END IF;
        v_safe := pg_catalog.lower(pg_catalog.regexp_replace(v_field,'[^a-zA-Z0-9_.-]+','_','g'));
        IF v_safe !~ '^[a-z][a-z0-9_.-]{0,127}$' THEN
            v_safe := 'legacy_' || pg_catalog.substring(pg_catalog.encode(
                pg_catalog.sha256(pg_catalog.convert_to(v_field,'UTF8')),'hex'),1,32);
        ELSE
            v_safe := pg_catalog.substring(v_safe,1,128);
        END IF;
        IF v_seen ? v_safe THEN
            CONTINUE;
        END IF;
        v_seen := v_seen || pg_catalog.jsonb_build_object(v_safe,true);
        v_result := v_result || pg_catalog.jsonb_build_array(pg_catalog.jsonb_strip_nulls(
            pg_catalog.jsonb_build_object(
                'field',v_safe,'classification','confidential',
                'before_sha256',CASE WHEN v_before IS NULL THEN NULL ELSE pg_catalog.encode(v_before,'hex') END,
                'after_sha256',CASE WHEN v_after IS NULL THEN NULL ELSE pg_catalog.encode(v_after,'hex') END)));
    END LOOP;
    RETURN v_result;
END;
$$;

CREATE TABLE audit_v1_adoption_runs (
    run_id uuid PRIMARY KEY,
    singleton boolean NOT NULL DEFAULT true UNIQUE CHECK (singleton),
    source_relation_oid oid,
    status text NOT NULL CHECK (status IN ('applying','completed')),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    adopted_count bigint NOT NULL CHECK (adopted_count BETWEEN 0 AND source_count),
    next_page bigint NOT NULL CHECK (next_page > 0),
    cursor_created_at timestamp without time zone NOT NULL,
    cursor_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    CHECK (pg_catalog.substring(run_id::text,15,1)='7'),
    CHECK ((status='applying' AND completed_at IS NULL) OR
           (status='completed' AND completed_at IS NOT NULL AND adopted_count=source_count))
);

CREATE TABLE audit_v1_adoption_mappings (
    source_id uuid PRIMARY KEY,
    run_id uuid NOT NULL REFERENCES audit_v1_adoption_runs(run_id) ON DELETE RESTRICT,
    target_entry_id uuid NOT NULL UNIQUE REFERENCES audit_entries(id) ON DELETE RESTRICT,
    target_owner_audit_id uuid NOT NULL UNIQUE,
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256)=32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256)=32),
    page_number bigint NOT NULL CHECK (page_number > 0),
    adopted_at timestamptz NOT NULL,
    CHECK (pg_catalog.substring(target_entry_id::text,15,1)='7'),
    CHECK (pg_catalog.substring(target_owner_audit_id::text,15,1)='7')
);

CREATE TABLE audit_v1_adoption_receipts (
    operation_id uuid PRIMARY KEY,
    run_id uuid NOT NULL REFERENCES audit_v1_adoption_runs(run_id) ON DELETE RESTRICT,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256)=32),
    page_number bigint NOT NULL CHECK (page_number > 0),
    row_count integer NOT NULL CHECK (row_count BETWEEN 0 AND 100),
    source_page_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_page_sha256)=32),
    target_page_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_page_sha256)=32),
    response_document jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(response_document)='object'),
    applied_at timestamptz NOT NULL,
    CHECK (pg_catalog.substring(operation_id::text,15,1)='7')
);

CREATE FUNCTION audit_guard_v1_adoption_evidence_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='AUDIT_V1_ADOPTION_EVIDENCE_IMMUTABLE';
END;
$$;

CREATE TRIGGER audit_v1_adoption_mappings_immutable
BEFORE UPDATE OR DELETE ON audit_v1_adoption_mappings
FOR EACH ROW EXECUTE FUNCTION audit_guard_v1_adoption_evidence_v1();
ALTER TABLE audit_v1_adoption_mappings ENABLE ALWAYS TRIGGER audit_v1_adoption_mappings_immutable;
CREATE TRIGGER audit_v1_adoption_receipts_immutable
BEFORE UPDATE OR DELETE ON audit_v1_adoption_receipts
FOR EACH ROW EXECUTE FUNCTION audit_guard_v1_adoption_evidence_v1();
ALTER TABLE audit_v1_adoption_receipts ENABLE ALWAYS TRIGGER audit_v1_adoption_receipts_immutable;
-- +goose StatementEnd
