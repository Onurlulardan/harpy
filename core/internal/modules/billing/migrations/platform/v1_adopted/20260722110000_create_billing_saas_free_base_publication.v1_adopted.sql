-- +goose Up
CREATE TABLE platform.billing_saas_free_base_publications (
    offering_key text PRIMARY KEY,
    offering_version_id uuid NOT NULL UNIQUE,
    revision bigint NOT NULL,
    offering_sha256 text NOT NULL,
    offering_document jsonb NOT NULL,
    effective_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_saas_free_base_publications_shape_check CHECK (
        offering_key = 'saas_free_base_v1'
        AND pg_catalog.uuid_extract_version(offering_version_id) = 7
        AND revision = 1
        AND offering_sha256 ~ '^[0-9a-f]{64}$'
        AND pg_catalog.jsonb_typeof(offering_document) = 'object'
        AND offering_sha256 = pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.convert_to(offering_document::text, 'UTF8')
        ), 'hex')
        AND pg_catalog.isfinite(effective_at)
    )
);

CREATE TRIGGER billing_saas_free_base_publications_reject_mutation
BEFORE UPDATE OR DELETE ON platform.billing_saas_free_base_publications
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
