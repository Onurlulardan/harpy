-- +goose Up
CREATE TABLE iam_actions (
    id uuid NOT NULL,
    action_key text COLLATE "C" NOT NULL,
    display_name text NOT NULL,
    description text,
    is_protected boolean NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_actions_pkey PRIMARY KEY (id),
    CONSTRAINT iam_actions_key_key UNIQUE (action_key),
    CONSTRAINT iam_actions_key_check CHECK (
        pg_catalog.char_length(action_key) BETWEEN 1 AND 100 AND
        action_key ~ '^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$'
    ),
    CONSTRAINT iam_actions_display_name_check CHECK (
        pg_catalog.char_length(display_name) BETWEEN 1 AND 255 AND
        display_name = pg_catalog.btrim(display_name)
    ),
    CONSTRAINT iam_actions_description_check CHECK (
        description IS NULL OR pg_catalog.char_length(description) <= 2000
    ),
    CONSTRAINT iam_actions_state_check CHECK (state IN ('active', 'retired')),
    CONSTRAINT iam_actions_revision_check CHECK (revision > 0),
    CONSTRAINT iam_actions_timestamp_order_check CHECK (updated_at >= created_at),
    CONSTRAINT iam_actions_retirement_check CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

CREATE TRIGGER iam_actions_revision_guard
BEFORE UPDATE ON iam_actions
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_actions ENABLE ALWAYS TRIGGER iam_actions_revision_guard;

CREATE TRIGGER iam_actions_immutable_guard
BEFORE UPDATE ON iam_actions
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'action_key', 'is_protected', 'created_at'
);
ALTER TABLE iam_actions ENABLE ALWAYS TRIGGER iam_actions_immutable_guard;

CREATE TRIGGER iam_actions_state_guard
BEFORE UPDATE ON iam_actions
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>retired');
ALTER TABLE iam_actions ENABLE ALWAYS TRIGGER iam_actions_state_guard;

CREATE TRIGGER iam_actions_terminal_guard
BEFORE UPDATE ON iam_actions
FOR EACH ROW EXECUTE FUNCTION iam_guard_rbac_terminal_v1();
ALTER TABLE iam_actions ENABLE ALWAYS TRIGGER iam_actions_terminal_guard;

CREATE TRIGGER iam_actions_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_actions
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_actions ENABLE ALWAYS TRIGGER iam_actions_removal_guard;
