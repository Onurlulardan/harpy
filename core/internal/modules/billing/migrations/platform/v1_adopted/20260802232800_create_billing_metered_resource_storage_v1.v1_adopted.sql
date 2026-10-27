-- +goose Up
CREATE TABLE platform.billing_metered_resource_receipts (
    operation_id uuid PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    mutation_kind text COLLATE "C" NOT NULL,
    owner_fact_id uuid NOT NULL,
    idempotency_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    feature_enabled boolean NOT NULL,
    allowed boolean NOT NULL,
    gauge_delta bigint NOT NULL,
    current_amount bigint NOT NULL,
    limit_amount bigint NOT NULL,
    unlimited boolean NOT NULL,
    usage_decision_id uuid,
    usage_entry_id uuid,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT billing_metered_resource_receipts_idempotency_key
        UNIQUE (tenant_id, actor_id, resource_kind, mutation_kind, idempotency_key),
    CONSTRAINT billing_metered_resource_receipts_authority_fk
        FOREIGN KEY (authority_id, tenant_id)
        REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resource_receipts_decision_fk
        FOREIGN KEY (usage_decision_id)
        REFERENCES platform.billing_entitlement_decisions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resource_receipts_entry_fk
        FOREIGN KEY (usage_entry_id)
        REFERENCES platform.billing_usage_entries (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resource_receipts_identity_check CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND pg_catalog.uuid_extract_version(owner_fact_id) = 7
        AND resource_kind IN ('api_endpoint','analytics_dashboard')
        AND mutation_kind IN (
            'create','update','publish','revalidate','delete','restore'
        )
        AND idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$'
        AND pg_catalog.octet_length(request_sha256) = 32
        AND current_amount >= 0 AND limit_amount >= 0
        AND (NOT unlimited OR limit_amount = 0)
        AND pg_catalog.isfinite(recorded_at)
    ),
    CONSTRAINT billing_metered_resource_receipts_outcome_check CHECK (
        (mutation_kind IN ('update','publish','revalidate')
         AND allowed = feature_enabled AND gauge_delta = 0
         AND current_amount = 0 AND limit_amount = 0 AND NOT unlimited)
        OR (mutation_kind IN ('create','restore')
         AND allowed = (
             feature_enabled AND (unlimited OR current_amount < limit_amount)
         )
         AND gauge_delta = CASE WHEN allowed THEN 1 ELSE 0 END)
        OR (mutation_kind = 'delete' AND allowed AND NOT feature_enabled
         AND gauge_delta IN (-1,0))
    ),
    CONSTRAINT billing_metered_resource_receipts_usage_check CHECK (
        (gauge_delta <> 0
         AND usage_decision_id IS NOT NULL AND usage_entry_id IS NOT NULL)
        OR (mutation_kind IN ('create','restore')
         AND feature_enabled AND NOT allowed AND gauge_delta = 0
         AND usage_decision_id IS NOT NULL AND usage_entry_id IS NULL)
        OR (gauge_delta = 0 AND NOT (
             mutation_kind IN ('create','restore')
             AND feature_enabled AND NOT allowed
         ) AND usage_decision_id IS NULL AND usage_entry_id IS NULL)
    )
);

CREATE TABLE platform.billing_metered_resources (
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    owner_fact_id uuid NOT NULL,
    counted_state text COLLATE "C" NOT NULL,
    origin_kind text COLLATE "C" NOT NULL,
    opening_evidence_sha256 bytea,
    opening_measured_at timestamptz,
    last_operation_id uuid NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT billing_metered_resources_pkey
        PRIMARY KEY (tenant_id, resource_kind, owner_fact_id),
    CONSTRAINT billing_metered_resources_authority_fk
        FOREIGN KEY (authority_id, tenant_id)
        REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_metered_resources_shape_check CHECK (
        pg_catalog.uuid_extract_version(owner_fact_id) = 7
        AND pg_catalog.uuid_extract_version(last_operation_id) = 7
        AND resource_kind IN ('api_endpoint','analytics_dashboard')
        AND counted_state IN ('active','released')
        AND origin_kind IN ('runtime','adopted')
        AND revision > 0
        AND pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND updated_at >= created_at
        AND (
            (origin_kind = 'runtime' AND opening_evidence_sha256 IS NULL
             AND opening_measured_at IS NULL)
            OR (origin_kind = 'adopted'
             AND pg_catalog.octet_length(opening_evidence_sha256) = 32
             AND pg_catalog.isfinite(opening_measured_at))
        )
    )
);

CREATE INDEX billing_metered_resources_state_idx
ON platform.billing_metered_resources (tenant_id, resource_kind, counted_state);

CREATE TRIGGER billing_metered_resource_receipts_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_metered_resource_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
CREATE TRIGGER billing_metered_resource_receipts_immutable_guard
BEFORE UPDATE OR DELETE ON platform.billing_metered_resource_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
CREATE TRIGGER billing_metered_resources_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_metered_resources
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
