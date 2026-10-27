-- +goose Up
ALTER TABLE platform.tenants
    ADD COLUMN lifecycle_revision bigint NOT NULL DEFAULT 1,
    ADD COLUMN offboarded_at timestamptz,
    ADD COLUMN purged_at timestamptz,
    ADD CONSTRAINT tenancy_tenants_lifecycle_revision_check
        CHECK (lifecycle_revision > 0),
    ADD CONSTRAINT tenancy_tenants_lifecycle_time_order_check
        CHECK (purged_at IS NULL OR
               (offboarded_at IS NOT NULL AND purged_at >= offboarded_at)),
    ADD CONSTRAINT tenancy_tenants_offboard_state_check
        CHECK (offboarded_at IS NULL OR
               (deleted_at IS NOT NULL AND status = 'SUSPENDED'));

CREATE TABLE platform.tenancy_customer_lifecycle_receipts (
    tenant_id uuid NOT NULL REFERENCES platform.tenants(id),
    control_tenant_id uuid NOT NULL REFERENCES platform.tenants(id),
    actor_id uuid NOT NULL,
    kind text NOT NULL CHECK (kind IN ('update','suspend','reactivate')),
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.length(request_sha256)=32),
    expected_revision bigint NOT NULL CHECK (expected_revision > 0),
    result_revision bigint NOT NULL CHECK (result_revision > 0),
    result_document jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(result_document)='object'),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id,idempotency_key),
    CHECK (tenant_id <> control_tenant_id),
    CHECK (idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$')
);

CREATE TRIGGER tenancy_customer_lifecycle_receipts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_customer_lifecycle_receipts
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();

CREATE TABLE platform.tenancy_customer_lifecycle_operations (
    operation_id uuid PRIMARY KEY DEFAULT pg_catalog.gen_random_uuid(),
    environment_id uuid NOT NULL,
    control_tenant_id uuid NOT NULL REFERENCES platform.tenants(id),
    target_tenant_id uuid NOT NULL REFERENCES platform.tenants(id),
    kind text NOT NULL CHECK (kind IN ('offboard','purge')),
    state text NOT NULL CHECK (state IN ('REQUESTED','RUNNING','SUCCEEDED','FAILED','UNKNOWN')),
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.length(request_sha256)=32),
    expected_revision bigint NOT NULL CHECK (expected_revision > 0),
    result_revision bigint NOT NULL CHECK (result_revision > 0),
    requested_by_actor_id uuid NOT NULL,
    retention_receipt_id uuid,
    backup_receipt_id uuid,
    contributor_receipt_id uuid,
    evidence_sha256 bytea,
    proof_verified_at timestamptz,
    worker_id uuid,
    leased_until timestamptz,
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    executor_receipt_id uuid,
    executor_receipt_sha256 bytea,
    failure_code text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (control_tenant_id,idempotency_key),
    CHECK (control_tenant_id <> target_tenant_id),
    CHECK (idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'),
    CHECK ((worker_id IS NULL) = (leased_until IS NULL)),
    CHECK ((executor_receipt_id IS NULL) = (executor_receipt_sha256 IS NULL)),
    CHECK (executor_receipt_sha256 IS NULL OR
           pg_catalog.length(executor_receipt_sha256)=32),
    CHECK ((kind='offboard' AND retention_receipt_id IS NULL AND
            backup_receipt_id IS NULL AND contributor_receipt_id IS NULL AND
            evidence_sha256 IS NULL AND proof_verified_at IS NULL) OR
           (kind='purge' AND retention_receipt_id IS NOT NULL AND
            backup_receipt_id IS NOT NULL AND contributor_receipt_id IS NOT NULL AND
            pg_catalog.length(evidence_sha256)=32 AND proof_verified_at IS NOT NULL))
);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_guard_customer_lifecycle_operation_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' OR OLD.state IN ('SUCCEEDED','FAILED','UNKNOWN') OR
       NEW.operation_id IS DISTINCT FROM OLD.operation_id OR
       NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
       NEW.control_tenant_id IS DISTINCT FROM OLD.control_tenant_id OR
       NEW.target_tenant_id IS DISTINCT FROM OLD.target_tenant_id OR
       NEW.kind IS DISTINCT FROM OLD.kind OR
       NEW.idempotency_key IS DISTINCT FROM OLD.idempotency_key OR
       NEW.request_sha256 IS DISTINCT FROM OLD.request_sha256 OR
       NEW.expected_revision IS DISTINCT FROM OLD.expected_revision OR
       NEW.requested_by_actor_id IS DISTINCT FROM OLD.requested_by_actor_id OR
       NEW.retention_receipt_id IS DISTINCT FROM OLD.retention_receipt_id OR
       NEW.backup_receipt_id IS DISTINCT FROM OLD.backup_receipt_id OR
       NEW.contributor_receipt_id IS DISTINCT FROM OLD.contributor_receipt_id OR
       NEW.evidence_sha256 IS DISTINCT FROM OLD.evidence_sha256 OR
       NEW.proof_verified_at IS DISTINCT FROM OLD.proof_verified_at OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       (OLD.state='REQUESTED' AND NEW.state<>'RUNNING') OR
       (OLD.state='RUNNING' AND NEW.state NOT IN
           ('RUNNING','SUCCEEDED','FAILED','UNKNOWN')) OR
       NEW.attempt_count<OLD.attempt_count THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Customer tenant lifecycle operation mutation denied';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER tenancy_customer_lifecycle_operations_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_customer_lifecycle_operations
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_guard_customer_lifecycle_operation_v1();

REVOKE ALL ON FUNCTION platform.tenancy_guard_customer_lifecycle_operation_v1() FROM PUBLIC;

CREATE INDEX tenancy_customer_lifecycle_operations_claim_idx
ON platform.tenancy_customer_lifecycle_operations(state,created_at,operation_id)
WHERE state IN ('REQUESTED','RUNNING');

REVOKE ALL ON TABLE platform.tenancy_customer_lifecycle_receipts
FROM PUBLIC,forgecrud_runtime;
REVOKE ALL ON TABLE platform.tenancy_customer_lifecycle_operations
FROM PUBLIC,forgecrud_runtime;
