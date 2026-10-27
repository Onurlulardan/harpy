-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION audit_field_changes_valid_v1(p_document jsonb) RETURNS boolean
LANGUAGE plpgsql IMMUTABLE SET search_path = pg_catalog AS $$
DECLARE
    v_item jsonb;
    v_classification text;
    v_before_digest text;
    v_after_digest text;
BEGIN
    IF p_document IS NULL OR pg_catalog.jsonb_typeof(p_document) <> 'array'
       OR pg_catalog.jsonb_array_length(p_document) > 100 THEN
        RETURN false;
    END IF;
    IF (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_array_elements(p_document)) <>
       (SELECT pg_catalog.count(DISTINCT value->>'field')
          FROM pg_catalog.jsonb_array_elements(p_document)) THEN
        RETURN false;
    END IF;
    FOR v_item IN SELECT value FROM pg_catalog.jsonb_array_elements(p_document) LOOP
        IF pg_catalog.jsonb_typeof(v_item) <> 'object'
           OR EXISTS (
                SELECT 1 FROM pg_catalog.jsonb_object_keys(v_item) AS keys(key)
                 WHERE keys.key NOT IN (
                    'field','classification','before','after','before_sha256','after_sha256'
                 )
           )
           OR (v_item->>'field') !~ '^[a-z][a-z0-9_.-]{0,127}$' THEN
            RETURN false;
        END IF;
        v_classification := v_item->>'classification';
        v_before_digest := v_item->>'before_sha256';
        v_after_digest := v_item->>'after_sha256';
        IF v_classification NOT IN ('public','internal','confidential','secret')
           OR (v_before_digest IS NOT NULL AND v_before_digest !~ '^[0-9a-f]{64}$')
           OR (v_after_digest IS NOT NULL AND v_after_digest !~ '^[0-9a-f]{64}$')
           OR (v_before_digest IS NULL AND v_after_digest IS NULL)
           OR (v_before_digest IS NOT NULL AND v_before_digest = v_after_digest) THEN
            RETURN false;
        END IF;
        IF v_classification IN ('confidential','secret') THEN
            IF v_item ? 'before' OR v_item ? 'after' THEN
                RETURN false;
            END IF;
        ELSE
            IF ((v_item ? 'before') <> (v_before_digest IS NOT NULL))
               OR ((v_item ? 'after') <> (v_after_digest IS NOT NULL))
               OR ((v_item ? 'before') AND (
                    pg_catalog.jsonb_typeof(v_item->'before') <> 'string'
                    OR pg_catalog.octet_length(v_item->>'before') > 2048
               ))
               OR ((v_item ? 'after') AND (
                    pg_catalog.jsonb_typeof(v_item->'after') <> 'string'
                    OR pg_catalog.octet_length(v_item->>'after') > 2048
               )) THEN
                RETURN false;
            END IF;
        END IF;
    END LOOP;
    RETURN true;
END;
$$;

CREATE TABLE audit_entries (
    id uuid PRIMARY KEY,
    owner_audit_id uuid NOT NULL UNIQUE,
    actor_id uuid NOT NULL,
    organization_id uuid,
    correlation_id uuid,
    owner_module text NOT NULL,
    operation_id text NOT NULL,
    entity_type text NOT NULL,
    entity_id text NOT NULL,
    form_slug text,
    action text NOT NULL,
    authorization_version bigint NOT NULL CHECK (authorization_version > 0),
    reason_code text,
    field_changes jsonb NOT NULL,
    event_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(event_sha256) = 32),
    occurred_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL,
    CHECK (pg_catalog.substring(id::text, 15, 1) = '7'),
    CHECK (pg_catalog.substring(owner_audit_id::text, 15, 1) = '7'),
    CHECK (actor_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (organization_id IS NULL OR organization_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (correlation_id IS NULL OR correlation_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (owner_module ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    CHECK (operation_id ~ '^[a-z][a-z0-9_.-]{0,127}$'),
    CHECK (entity_type ~ '^[a-z][a-z0-9_.-]{0,127}$'),
    CHECK (pg_catalog.octet_length(entity_id) BETWEEN 1 AND 200),
    CHECK (form_slug IS NULL OR form_slug ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    CHECK (action IN ('create','update','delete','transition','publish','execute',
                      'grant','revoke','restore','purge')),
    CHECK (reason_code IS NULL OR reason_code ~ '^[a-z][a-z0-9_.-]{0,127}$'),
    CHECK (audit_field_changes_valid_v1(field_changes)),
    CHECK (pg_catalog.isfinite(occurred_at) AND pg_catalog.isfinite(recorded_at)),
    CHECK (recorded_at >= occurred_at)
);

CREATE INDEX audit_entries_time_idx ON audit_entries(occurred_at DESC, id DESC);
CREATE INDEX audit_entries_entity_idx
    ON audit_entries(entity_type, entity_id, occurred_at DESC, id DESC);
CREATE INDEX audit_entries_form_idx
    ON audit_entries(form_slug, entity_id, occurred_at DESC, id DESC)
    WHERE form_slug IS NOT NULL;
CREATE INDEX audit_entries_actor_idx ON audit_entries(actor_id, occurred_at DESC, id DESC);
CREATE INDEX audit_entries_org_idx
    ON audit_entries(organization_id, occurred_at DESC, id DESC)
    WHERE organization_id IS NOT NULL;
CREATE INDEX audit_entries_correlation_idx
    ON audit_entries(correlation_id) WHERE correlation_id IS NOT NULL;

CREATE FUNCTION audit_guard_entry_immutable_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'AUDIT_ENTRY_IMMUTABLE';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER audit_entries_immutable_guard
BEFORE UPDATE OR DELETE ON audit_entries
FOR EACH ROW EXECUTE FUNCTION audit_guard_entry_immutable_v1();
ALTER TABLE audit_entries ENABLE ALWAYS TRIGGER audit_entries_immutable_guard;
