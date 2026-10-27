-- +goose Up
ALTER TABLE documents_links
    DROP CONSTRAINT documents_links_source_owner_check,
    ADD CONSTRAINT documents_links_source_owner_check
    CHECK (source_owner IN ('FORMS','WORKFLOW','WORK','CATALOG','IAM','AI'));

CREATE TABLE documents_knowledge_file_leases (
    lease_id uuid PRIMARY KEY,
    upload_id uuid NOT NULL UNIQUE,
    actor_id uuid NOT NULL,
    item_id uuid NOT NULL UNIQUE,
    version_id uuid NOT NULL UNIQUE,
    reference_id uuid UNIQUE,
    source_id uuid,
    expected_plaintext_bytes bigint NOT NULL
        CHECK (expected_plaintext_bytes BETWEEN 1 AND 20971520),
    expected_plaintext_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(expected_plaintext_sha256) = 32),
    filename text NOT NULL
        CHECK (pg_catalog.octet_length(filename) BETWEEN 1 AND 255),
    declared_mime text NOT NULL
        CHECK (pg_catalog.octet_length(declared_mime) BETWEEN 1 AND 255),
    detected_mime text,
    scan_signature_revision text,
    scanned_at timestamptz,
    lifecycle_state text NOT NULL
        CHECK (lifecycle_state IN ('PENDING','STAGED','ADOPTED','RELEASED','EXPIRED')),
    expires_at timestamptz NOT NULL,
    adopted_at timestamptz,
    released_at timestamptz,
    expired_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT documents_knowledge_leases_upload_fk FOREIGN KEY (upload_id)
        REFERENCES documents_uploads(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT documents_knowledge_leases_reference_fk FOREIGN KEY (reference_id)
        REFERENCES documents_links(id) ON UPDATE NO ACTION ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CHECK (lease_id = upload_id),
    CHECK (pg_catalog.isfinite(expires_at) AND pg_catalog.isfinite(created_at)
        AND pg_catalog.isfinite(updated_at)),
    CHECK ((detected_mime IS NULL) = (scan_signature_revision IS NULL)
        AND (detected_mime IS NULL) = (scanned_at IS NULL)),
    CHECK ((reference_id IS NULL) = (source_id IS NULL)),
    CHECK (
        lifecycle_state = 'PENDING' AND detected_mime IS NULL
            AND reference_id IS NULL AND adopted_at IS NULL
            AND released_at IS NULL AND expired_at IS NULL
        OR lifecycle_state = 'STAGED' AND detected_mime IS NOT NULL
            AND reference_id IS NULL AND adopted_at IS NULL
            AND released_at IS NULL AND expired_at IS NULL
        OR lifecycle_state = 'ADOPTED' AND detected_mime IS NOT NULL
            AND reference_id IS NOT NULL AND adopted_at IS NOT NULL
            AND released_at IS NULL AND expired_at IS NULL
        OR lifecycle_state = 'RELEASED' AND detected_mime IS NOT NULL
            AND reference_id IS NOT NULL AND adopted_at IS NOT NULL
            AND released_at IS NOT NULL AND expired_at IS NULL
        OR lifecycle_state = 'EXPIRED' AND reference_id IS NULL
            AND adopted_at IS NULL AND released_at IS NULL
            AND expired_at IS NOT NULL
    )
);

CREATE INDEX documents_knowledge_file_lease_expiry_idx
    ON documents_knowledge_file_leases(expires_at, lease_id)
    WHERE lifecycle_state IN ('PENDING','STAGED');

-- +goose StatementBegin
CREATE FUNCTION documents_guard_knowledge_file_lease_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    IF NEW.lease_id <> OLD.lease_id OR NEW.upload_id <> OLD.upload_id
       OR NEW.actor_id <> OLD.actor_id OR NEW.item_id <> OLD.item_id
       OR NEW.version_id <> OLD.version_id
       OR NEW.expected_plaintext_bytes <> OLD.expected_plaintext_bytes
       OR NEW.expected_plaintext_sha256 <> OLD.expected_plaintext_sha256
       OR NEW.filename <> OLD.filename OR NEW.declared_mime <> OLD.declared_mime
       OR NEW.expires_at <> OLD.expires_at OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'documents knowledge-file lease identity is immutable';
    END IF;
    IF NOT (
        NEW.lifecycle_state = OLD.lifecycle_state
        OR OLD.lifecycle_state = 'PENDING' AND NEW.lifecycle_state IN ('STAGED','EXPIRED')
        OR OLD.lifecycle_state = 'STAGED' AND NEW.lifecycle_state IN ('ADOPTED','EXPIRED')
        OR OLD.lifecycle_state = 'ADOPTED' AND NEW.lifecycle_state = 'RELEASED'
    ) THEN
        RAISE EXCEPTION 'documents knowledge-file lease transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_guard_knowledge_file_lease_v1() FROM PUBLIC;
CREATE TRIGGER documents_knowledge_file_lease_transition_guard
BEFORE UPDATE ON documents_knowledge_file_leases
FOR EACH ROW EXECUTE FUNCTION documents_guard_knowledge_file_lease_v1();
ALTER TABLE documents_knowledge_file_leases
    ENABLE ALWAYS TRIGGER documents_knowledge_file_lease_transition_guard;

REVOKE ALL ON TABLE documents_knowledge_file_leases FROM PUBLIC;
GRANT SELECT, INSERT ON documents_knowledge_file_leases TO harpy_runtime;
GRANT UPDATE (
    reference_id, source_id, detected_mime, scan_signature_revision, scanned_at,
    lifecycle_state, adopted_at, released_at, expired_at, updated_at
) ON documents_knowledge_file_leases TO harpy_runtime;

