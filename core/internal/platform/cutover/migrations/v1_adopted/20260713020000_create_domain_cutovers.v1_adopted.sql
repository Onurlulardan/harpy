-- +goose Up
CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    state text NOT NULL,
    writer text NOT NULL,
    generation bigint NOT NULL,
    requested_by uuid NOT NULL,
    approved_by uuid,
    reason text NOT NULL,
    change_reference text NOT NULL,
    acceptance_digest text,
    changed_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (scope_kind, scope_id, domain),
    CONSTRAINT domain_cutovers_scope_kind_check
        CHECK (scope_kind IN ('global', 'tenant')),
    CONSTRAINT domain_cutovers_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'forms', 'iam',
            'notifications', 'work', 'workflow'
        )),
    CONSTRAINT domain_cutovers_global_check
        CHECK (scope_kind <> 'global' OR domain = 'billing'),
    CONSTRAINT domain_cutovers_state_check
        CHECK (state IN ('v1', 'shadow', 'quiescing', 'v2', 'fenced')),
    CONSTRAINT domain_cutovers_writer_state_check
        CHECK (
            (state IN ('v1', 'shadow') AND writer = 'v1') OR
            (state = 'v2' AND writer = 'v2') OR
            (state IN ('quiescing', 'fenced') AND writer = 'none')
        ),
    CONSTRAINT domain_cutovers_generation_check CHECK (generation > 0),
    CONSTRAINT domain_cutovers_approval_check
        CHECK (approved_by IS NULL OR approved_by <> requested_by),
    CONSTRAINT domain_cutovers_reason_check
        CHECK (pg_catalog.octet_length(reason) BETWEEN 1 AND 2048),
    CONSTRAINT domain_cutovers_reference_check
        CHECK (pg_catalog.octet_length(change_reference) BETWEEN 1 AND 256),
    CONSTRAINT domain_cutovers_acceptance_check
        CHECK (acceptance_digest IS NULL OR acceptance_digest ~ '^[0-9a-f]{64}$')
);

CREATE INDEX domain_cutovers_rollout_idx
    ON platform.domain_cutovers (state, writer, domain, scope_kind);
