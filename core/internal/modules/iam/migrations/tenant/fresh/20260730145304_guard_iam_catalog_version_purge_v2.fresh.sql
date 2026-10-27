-- +goose Up
CREATE TABLE iam_catalog_purge_contexts (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    opened_at timestamptz NOT NULL,
    PRIMARY KEY (backend_pid,transaction_id,component_id),
    UNIQUE (operation_id,component_id),
    FOREIGN KEY (component_id)
        REFERENCES iam_catalog_components(component_id) ON DELETE RESTRICT,
    CHECK (backend_pid > 0),
    CHECK (pg_catalog.uuid_extract_version(operation_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.isfinite(opened_at)),
    CHECK (date_trunc('microseconds',opened_at) = opened_at)
);

-- +goose StatementBegin
CREATE FUNCTION iam_catalog_purge_context_matches_v2(p_component_id uuid)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM iam_catalog_purge_contexts AS context
        WHERE context.backend_pid=pg_catalog.pg_backend_pid()
          AND context.transaction_id=pg_catalog.pg_current_xact_id()
          AND context.component_id=p_component_id
    );
END;
$function$;

CREATE FUNCTION iam_catalog_version_evidence_guard_v3()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP='DELETE' AND
       iam_catalog_purge_context_matches_v2(OLD.component_id) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='IAM Catalog version evidence is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_purge_context_matches_v2(uuid)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_catalog_version_evidence_guard_v3()
    FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_catalog_purge_context_matches_v2(uuid)
    FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_catalog_version_evidence_guard_v3()
    FROM harpy_runtime;

CREATE OR REPLACE TRIGGER iam_catalog_version_evidence_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_evidence
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_guard_v3();
CREATE OR REPLACE TRIGGER iam_catalog_version_resources_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_resources
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_guard_v3();
CREATE OR REPLACE TRIGGER iam_catalog_version_grants_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_grants
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_guard_v3();
CREATE OR REPLACE TRIGGER iam_catalog_version_provisions_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_provisions
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_guard_v3();
CREATE OR REPLACE TRIGGER iam_catalog_version_operational_resources_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_operational_resources
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_guard_v3();
CREATE OR REPLACE TRIGGER iam_catalog_version_operational_grants_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_operational_grants
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_guard_v3();
