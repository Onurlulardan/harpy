-- +goose Up
CREATE TABLE platform.cutover_job_bindings (
    job_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    cutover_run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    group_id text NOT NULL CHECK (group_id = 'documents'),
    group_generation bigint NOT NULL CHECK (group_generation > 0),
    domain text NOT NULL CHECK (domain = 'documents'),
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    lease_manifest_sha256 text NOT NULL
        CHECK (lease_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    adoption_run_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    job_kind text NOT NULL CHECK (job_kind = 'documents.adoption.copy'),
    payload_revision integer NOT NULL CHECK (payload_revision = 1),
    payload_ref text NOT NULL
        CHECK (pg_catalog.octet_length(payload_ref) BETWEEN 1 AND 2048),
    payload_sha256 text NOT NULL CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    effect_idempotency_key text NOT NULL
        CHECK (pg_catalog.octet_length(effect_idempotency_key) BETWEEN 1 AND 512),
    enqueued_at timestamptz NOT NULL,
    CONSTRAINT cutover_job_bindings_operation_fk
        FOREIGN KEY (operation_id, cutover_run_id)
        REFERENCES platform.cutover_operations (operation_id, run_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_job_bindings_job_fk FOREIGN KEY (job_id)
        REFERENCES platform.jobs (id) ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT cutover_job_bindings_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT cutover_job_bindings_source_unique UNIQUE (
        operation_id, cutover_run_id, adoption_run_id, source_version_id
    )
);

CREATE INDEX cutover_job_bindings_authority_idx
    ON platform.cutover_job_bindings (
        tenant_id, group_id, group_generation, domain, cutover_generation
    );

CREATE TRIGGER cutover_job_bindings_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_job_bindings
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.cutover_job_bindings
    ENABLE ALWAYS TRIGGER cutover_job_bindings_immutable_guard;
