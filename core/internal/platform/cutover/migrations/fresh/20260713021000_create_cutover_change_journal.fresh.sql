-- +goose Up
CREATE TABLE platform.cutover_change_journal (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    cutover_generation bigint NOT NULL,
    source_xid xid8 NOT NULL DEFAULT pg_catalog.pg_current_xact_id(),
    source_session_user text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (run_id, scope_kind, scope_id, domain, source_xid),
    CONSTRAINT cutover_change_journal_operation_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.cutover_operations (operation_id, run_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_change_journal_domain_fk
        FOREIGN KEY (scope_kind, scope_id, domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_change_journal_scope_check
        CHECK (scope_kind IN ('global', 'tenant')),
    CONSTRAINT cutover_change_journal_generation_check CHECK (cutover_generation > 0),
    CONSTRAINT cutover_change_journal_session_check
        CHECK (source_session_user = 'harpy_v1_runtime')
);

CREATE INDEX cutover_change_journal_operation_idx
    ON platform.cutover_change_journal (operation_id, run_id, created_at);
