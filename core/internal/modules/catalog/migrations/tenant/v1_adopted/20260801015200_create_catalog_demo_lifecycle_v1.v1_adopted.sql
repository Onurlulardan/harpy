-- +goose Up
ALTER TABLE catalog_operation_runs
    DROP CONSTRAINT catalog_operation_runs_kind_check;
ALTER TABLE catalog_operation_runs
    ADD CONSTRAINT catalog_operation_runs_kind_check CHECK (kind IN (
        'INSTALL','UPGRADE','RETRY','RESTORE','RETIRE','PURGE',
        'DEMO_INSTALL','DEMO_REMOVE'
    ));

CREATE TABLE catalog_demo_states (
    installation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    package_id uuid NOT NULL,
    release_id uuid NOT NULL,
    source_run_id uuid NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    state text NOT NULL CHECK (state IN ('INSTALLED','ABSENT')),
    record_ids jsonb NOT NULL CHECK (
        jsonb_typeof(record_ids) = 'array'
        AND jsonb_array_length(record_ids) BETWEEN 0 AND 10000
    ),
    last_receipt_id uuid,
    revision bigint NOT NULL CHECK (revision > 0),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (installation_id, component_id),
    CONSTRAINT catalog_demo_states_binding_fk
        FOREIGN KEY (installation_id, component_id)
        REFERENCES catalog_component_bindings(installation_id, component_id)
        ON DELETE RESTRICT,
    CONSTRAINT catalog_demo_states_receipt_fk
        FOREIGN KEY (last_receipt_id)
        REFERENCES catalog_demo_receipts(receipt_id)
        ON DELETE RESTRICT,
    CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(release_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(source_run_id) = 7),
    CHECK (date_trunc('microseconds', updated_at) = updated_at)
);

INSERT INTO catalog_demo_states (
    installation_id, component_id, package_id, release_id, source_run_id,
    content_sha256, state, record_ids, last_receipt_id, revision, updated_at
)
SELECT binding.installation_id, head.component_id, head.package_id,
       head.release_id, head.forms_run_id, head.content_sha256,
       'INSTALLED', '[]'::jsonb, NULL, 1, head.activated_at
FROM catalog_demo_component_heads AS head
JOIN catalog_component_bindings AS binding
  ON binding.component_id = head.component_id
 AND binding.owner_module = 'catalog'
 AND binding.owner_contract = 'catalog.demo-data.v1'
ON CONFLICT (installation_id, component_id) DO NOTHING;

CREATE TABLE catalog_demo_operation_plans (
    operation_id uuid NOT NULL,
    installation_id uuid NOT NULL,
    action text NOT NULL CHECK (action IN ('INSTALL','REMOVE')),
    component_id uuid NOT NULL,
    source_run_id uuid NOT NULL,
    record_ids jsonb NOT NULL CHECK (
        jsonb_typeof(record_ids) = 'array'
        AND jsonb_array_length(record_ids) BETWEEN 0 AND 10000
    ),
    expected_plan_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, component_id),
    CONSTRAINT catalog_demo_operation_plans_operation_fk
        FOREIGN KEY (operation_id) REFERENCES catalog_operation_runs(operation_id)
        ON DELETE RESTRICT,
    CONSTRAINT catalog_demo_operation_plans_binding_fk
        FOREIGN KEY (installation_id, component_id)
        REFERENCES catalog_component_bindings(installation_id, component_id)
        ON DELETE RESTRICT,
    CHECK (
        (action = 'INSTALL' AND expected_plan_sha256 = '')
        OR (action = 'REMOVE' AND expected_plan_sha256 ~ '^[0-9a-f]{64}$')
    ),
    CHECK (pg_catalog.uuid_extract_version(operation_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(component_id) = 7),
    CHECK (pg_catalog.uuid_extract_version(source_run_id) = 7),
    CHECK (date_trunc('microseconds', created_at) = created_at)
);

ALTER TABLE catalog_demo_receipts
    ADD CONSTRAINT catalog_demo_receipts_operation_fk
    FOREIGN KEY (operation_id) REFERENCES catalog_operation_runs(operation_id)
    ON DELETE RESTRICT NOT VALID;
ALTER TABLE catalog_demo_receipts
    VALIDATE CONSTRAINT catalog_demo_receipts_operation_fk;

GRANT SELECT, INSERT ON TABLE catalog_demo_operation_plans TO harpy_runtime;
GRANT SELECT ON TABLE catalog_demo_states TO harpy_runtime;
GRANT SELECT ON TABLE catalog_demo_payloads TO harpy_runtime;
GRANT INSERT (installation_id, component_id, package_id, release_id,
    source_run_id, content_sha256, state, record_ids, last_receipt_id,
    revision, updated_at) ON catalog_demo_states TO harpy_runtime;
GRANT UPDATE (source_run_id, state, record_ids, last_receipt_id,
    revision, updated_at) ON catalog_demo_states TO harpy_runtime;
GRANT SELECT ON TABLE catalog_demo_receipts TO harpy_runtime;
GRANT INSERT (receipt_id, installation_id, operation_id, component_id,
    action, stable_ids, content_sha256, actor_id, idempotency_key,
    completed_at) ON catalog_demo_receipts TO harpy_runtime;
