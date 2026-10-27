-- +goose Up
CREATE TABLE platform.cutover_transition_events (
    event_id uuid NOT NULL,
    request_event_id uuid,
    target_kind text NOT NULL,
    event_kind text NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text,
    previous_state text NOT NULL,
    new_state text NOT NULL,
    previous_writer text NOT NULL,
    new_writer text NOT NULL,
    previous_generation bigint NOT NULL,
    new_generation bigint NOT NULL,
    actor_id uuid NOT NULL,
    requested_by uuid NOT NULL,
    approved_by uuid,
    reason text NOT NULL,
    change_reference text NOT NULL,
    acceptance_digest text,
    event_xid xid8 NOT NULL DEFAULT pg_catalog.pg_current_xact_id(),
    occurred_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (event_id),
    CONSTRAINT cutover_transition_events_request_fk
        FOREIGN KEY (request_event_id)
        REFERENCES platform.cutover_transition_events (event_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_transition_events_group_fk
        FOREIGN KEY (scope_kind, scope_id, group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_transition_events_domain_fk
        FOREIGN KEY (scope_kind, scope_id, domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_transition_events_group_member_fk
        FOREIGN KEY (scope_kind, scope_id, group_id, domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        )
        ON DELETE RESTRICT,
    CONSTRAINT cutover_transition_events_target_check
        CHECK (
            (target_kind = 'group' AND domain IS NULL) OR
            (target_kind = 'domain' AND domain IS NOT NULL)
        ),
    CONSTRAINT cutover_transition_events_kind_check
        CHECK (event_kind IN (
            'request', 'approval', 'fence', 'quiesce', 'switch',
            'rollback', 'blocked', 'closure'
        )),
    CONSTRAINT cutover_transition_events_scope_check
        CHECK (scope_kind IN ('global', 'tenant')),
    CONSTRAINT cutover_transition_events_group_check
        CHECK (group_id ~ '^[a-z][a-z0-9_]{0,62}$'),
    CONSTRAINT cutover_transition_events_domain_check
        CHECK (
            domain IS NULL OR domain IN (
                'analytics', 'billing', 'catalog', 'documents', 'iam',
                'notifications', 'work', 'workflow'
            )
        ),
    CONSTRAINT cutover_transition_events_state_check
        CHECK (
            previous_state IN ('v1', 'shadow', 'quiescing', 'v2', 'fenced') AND
            new_state IN ('v1', 'shadow', 'quiescing', 'v2', 'fenced')
        ),
    CONSTRAINT cutover_transition_events_previous_writer_check
        CHECK (
            (previous_state IN ('v1', 'shadow') AND previous_writer = 'v1') OR
            (previous_state = 'v2' AND previous_writer = 'v2') OR
            (previous_state IN ('quiescing', 'fenced') AND previous_writer = 'none')
        ),
    CONSTRAINT cutover_transition_events_new_writer_check
        CHECK (
            (new_state IN ('v1', 'shadow') AND new_writer = 'v1') OR
            (new_state = 'v2' AND new_writer = 'v2') OR
            (new_state IN ('quiescing', 'fenced') AND new_writer = 'none')
        ),
    CONSTRAINT cutover_transition_events_generation_check
        CHECK (previous_generation > 0 AND new_generation >= previous_generation),
    CONSTRAINT cutover_transition_events_principal_check
        CHECK (approved_by IS NULL OR approved_by <> requested_by),
    CONSTRAINT cutover_transition_events_approval_check
        CHECK (
            event_kind <> 'approval' OR
            (request_event_id IS NOT NULL AND approved_by = actor_id AND requested_by <> actor_id)
        ),
    CONSTRAINT cutover_transition_events_request_check
        CHECK (event_kind <> 'request' OR request_event_id IS NULL),
    CONSTRAINT cutover_transition_events_reason_check
        CHECK (pg_catalog.octet_length(reason) BETWEEN 1 AND 2048),
    CONSTRAINT cutover_transition_events_reference_check
        CHECK (pg_catalog.octet_length(change_reference) BETWEEN 1 AND 256),
    CONSTRAINT cutover_transition_events_acceptance_check
        CHECK (acceptance_digest IS NULL OR acceptance_digest ~ '^[0-9a-f]{64}$')
);

CREATE INDEX cutover_transition_events_target_idx
    ON platform.cutover_transition_events (
        scope_kind, scope_id, group_id, domain, new_generation, occurred_at
    );

CREATE INDEX cutover_transition_events_xid_idx
    ON platform.cutover_transition_events (event_xid, target_kind);
