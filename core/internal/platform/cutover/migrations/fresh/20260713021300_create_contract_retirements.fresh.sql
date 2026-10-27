-- +goose Up
CREATE TABLE platform.contract_retirements (
    contract_key text NOT NULL,
    owner_module text NOT NULL,
    domain text NOT NULL,
    state text NOT NULL,
    remaining_scope_count bigint NOT NULL,
    hit_count bigint NOT NULL,
    last_hit_at timestamptz,
    zero_hit_since timestamptz,
    eligible_at timestamptz,
    retired_at timestamptz,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (contract_key),
    CONSTRAINT contract_retirements_key_check
        CHECK (contract_key ~ '^[a-z][a-z0-9_.-]{0,127}$'),
    CONSTRAINT contract_retirements_owner_check
        CHECK (owner_module ~ '^[a-z][a-z0-9_]{0,62}$'),
    CONSTRAINT contract_retirements_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'iam',
            'notifications', 'work', 'workflow'
        )),
    CONSTRAINT contract_retirements_state_check
        CHECK (state IN ('active', 'eligible', 'retired')),
    CONSTRAINT contract_retirements_count_check
        CHECK (remaining_scope_count >= 0 AND hit_count >= 0 AND revision > 0),
    CONSTRAINT contract_retirements_eligible_check
        CHECK (
            state = 'active' OR
            (remaining_scope_count = 0 AND zero_hit_since IS NOT NULL AND eligible_at IS NOT NULL)
        ),
    CONSTRAINT contract_retirements_retired_check
        CHECK (
            (state = 'retired' AND retired_at IS NOT NULL) OR
            (state <> 'retired' AND retired_at IS NULL)
        ),
    CONSTRAINT contract_retirements_time_check
        CHECK (
            (last_hit_at IS NULL OR last_hit_at <= updated_at) AND
            (zero_hit_since IS NULL OR zero_hit_since <= updated_at) AND
            (eligible_at IS NULL OR eligible_at <= updated_at) AND
            (retired_at IS NULL OR retired_at <= updated_at)
        )
);

CREATE INDEX contract_retirements_gate_idx
    ON platform.contract_retirements (state, remaining_scope_count, zero_hit_since);
