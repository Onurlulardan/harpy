-- +goose Up
CREATE TABLE iam_permission_implications (
    id uuid NOT NULL,
    implication_kind text NOT NULL,
    source_permission_id uuid NOT NULL,
    source_match_kind text NOT NULL,
    implied_permission_id uuid NOT NULL,
    implied_match_kind text NOT NULL,
    retirement_contract_key text COLLATE "C" NOT NULL,
    state text NOT NULL,
    hit_count bigint NOT NULL,
    last_hit_at timestamptz,
    retirement_receipt_id uuid,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_permission_implications_pkey PRIMARY KEY (id),
    CONSTRAINT iam_permission_implications_edge_key
        UNIQUE (implication_kind, source_permission_id, implied_permission_id),
    CONSTRAINT iam_permission_implications_source_fk
        FOREIGN KEY (source_permission_id, source_match_kind)
        REFERENCES iam_permissions (id, match_kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_implications_implied_fk
        FOREIGN KEY (implied_permission_id, implied_match_kind)
        REFERENCES iam_permissions (id, match_kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_implications_kind_check
        CHECK (implication_kind = 'v1_form_parent_v1'),
    CONSTRAINT iam_permission_implications_exact_check
        CHECK (source_match_kind = 'exact' AND implied_match_kind = 'exact'),
    CONSTRAINT iam_permission_implications_distinct_check
        CHECK (source_permission_id <> implied_permission_id),
    CONSTRAINT iam_permission_implications_contract_key_check CHECK (
        pg_catalog.char_length(retirement_contract_key) BETWEEN 1 AND 100 AND
        retirement_contract_key ~ '^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$'
    ),
    CONSTRAINT iam_permission_implications_state_check
        CHECK (state IN ('active', 'retired')),
    CONSTRAINT iam_permission_implications_hit_count_check CHECK (hit_count >= 0),
    CONSTRAINT iam_permission_implications_hit_time_check CHECK (
        (hit_count = 0 AND last_hit_at IS NULL) OR
        (hit_count > 0 AND last_hit_at IS NOT NULL AND
            last_hit_at >= created_at AND last_hit_at <= updated_at)
    ),
    CONSTRAINT iam_permission_implications_revision_check CHECK (revision > 0),
    CONSTRAINT iam_permission_implications_timestamp_order_check
        CHECK (updated_at >= created_at),
    CONSTRAINT iam_permission_implications_retirement_check CHECK (
        (state = 'active' AND retired_at IS NULL AND retirement_receipt_id IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND
            retired_at >= updated_at AND retirement_receipt_id IS NOT NULL)
    )
);

CREATE INDEX iam_permission_implications_implied_idx
    ON iam_permission_implications (implied_permission_id, state);

-- +goose StatementBegin
CREATE FUNCTION iam_guard_permission_implication_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    eligible boolean;
BEGIN
    IF TG_OP NOT IN ('INSERT', 'UPDATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM permission implication guard invocation is invalid';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.hit_count < OLD.hit_count OR
        (OLD.last_hit_at IS NOT NULL AND NEW.last_hit_at < OLD.last_hit_at)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_permission_implications_hit_guard',
            MESSAGE = 'IAM implication hit evidence cannot decrease';
    END IF;
    IF NEW.state = 'active' THEN
        SELECT pg_catalog.count(*) = 2 INTO eligible
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        JOIN iam_actions AS action ON action.id = permission.action_id
        WHERE permission.id IN (NEW.source_permission_id, NEW.implied_permission_id)
          AND permission.match_kind = 'exact' AND permission.state = 'active'
          AND resource.state = 'active' AND action.state = 'active';
        IF eligible IS DISTINCT FROM true THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                CONSTRAINT = 'iam_permission_implications_active_guard',
                MESSAGE = 'Active implication requires two active exact permissions';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_permission_implication_v1() FROM PUBLIC;

CREATE TRIGGER iam_permission_implications_revision_guard
BEFORE UPDATE ON iam_permission_implications
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_permission_implications
    ENABLE ALWAYS TRIGGER iam_permission_implications_revision_guard;

CREATE TRIGGER iam_permission_implications_immutable_guard
BEFORE UPDATE ON iam_permission_implications
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'implication_kind', 'source_permission_id', 'source_match_kind',
    'implied_permission_id', 'implied_match_kind',
    'retirement_contract_key', 'created_at'
);
ALTER TABLE iam_permission_implications
    ENABLE ALWAYS TRIGGER iam_permission_implications_immutable_guard;

CREATE TRIGGER iam_permission_implications_state_guard
BEFORE UPDATE ON iam_permission_implications
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>retired');
ALTER TABLE iam_permission_implications
    ENABLE ALWAYS TRIGGER iam_permission_implications_state_guard;

CREATE TRIGGER iam_permission_implications_terminal_guard
BEFORE UPDATE ON iam_permission_implications
FOR EACH ROW EXECUTE FUNCTION iam_guard_rbac_terminal_v1();
ALTER TABLE iam_permission_implications
    ENABLE ALWAYS TRIGGER iam_permission_implications_terminal_guard;

CREATE TRIGGER iam_permission_implications_evidence_guard
BEFORE INSERT OR UPDATE ON iam_permission_implications
FOR EACH ROW EXECUTE FUNCTION iam_guard_permission_implication_v1();
ALTER TABLE iam_permission_implications
    ENABLE ALWAYS TRIGGER iam_permission_implications_evidence_guard;

CREATE TRIGGER iam_permission_implications_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_permission_implications
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permission_implications
    ENABLE ALWAYS TRIGGER iam_permission_implications_removal_guard;
