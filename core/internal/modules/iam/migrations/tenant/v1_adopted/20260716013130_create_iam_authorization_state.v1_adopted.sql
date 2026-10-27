-- +goose Up
CREATE TABLE iam_authorization_clock (
    singleton boolean NOT NULL,
    clock_value bigint NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_authorization_clock_pkey PRIMARY KEY (singleton),
    CONSTRAINT iam_authorization_clock_singleton_check CHECK (singleton),
    CONSTRAINT iam_authorization_clock_value_check CHECK (clock_value >= 0),
    CONSTRAINT iam_authorization_clock_revision_check CHECK (revision > 0),
    CONSTRAINT iam_authorization_clock_timestamp_order_check
        CHECK (updated_at >= created_at)
);

CREATE TABLE iam_authorization_operation_receipts (
    operation_id uuid NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    clock_value bigint NOT NULL,
    result_code text COLLATE "C" NOT NULL,
    result_id uuid NOT NULL,
    result_revision bigint NOT NULL,
    result_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_authorization_operation_receipts_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_authorization_operation_receipts_command_key_check CHECK (
        pg_catalog.char_length(command_key) BETWEEN 1 AND 100 AND
        command_key ~ '^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$'
    ),
    CONSTRAINT iam_authorization_operation_receipts_request_hash_check
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    CONSTRAINT iam_authorization_operation_receipts_clock_check
        CHECK (clock_value > 0),
    CONSTRAINT iam_authorization_operation_receipts_result_code_check CHECK (
        pg_catalog.char_length(result_code) BETWEEN 1 AND 100 AND
        result_code ~ '^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$'
    ),
    CONSTRAINT iam_authorization_operation_receipts_result_revision_check
        CHECK (result_revision > 0),
    CONSTRAINT iam_authorization_operation_receipts_result_hash_check
        CHECK (pg_catalog.octet_length(result_sha256) = 32)
);

-- +goose StatementBegin
CREATE FUNCTION iam_guard_authorization_clock_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       NEW.clock_value IS DISTINCT FROM OLD.clock_value + 1 OR
       NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_authorization_clock_step_guard',
            MESSAGE = 'IAM authorization clock must advance exactly once';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_authorization_clock_v1() FROM PUBLIC;

CREATE TRIGGER iam_authorization_clock_revision_guard
BEFORE UPDATE ON iam_authorization_clock
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_authorization_clock
    ENABLE ALWAYS TRIGGER iam_authorization_clock_revision_guard;

CREATE TRIGGER iam_authorization_clock_immutable_guard
BEFORE UPDATE ON iam_authorization_clock
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'singleton', 'created_at'
);
ALTER TABLE iam_authorization_clock
    ENABLE ALWAYS TRIGGER iam_authorization_clock_immutable_guard;

CREATE TRIGGER iam_authorization_clock_step_guard
BEFORE UPDATE ON iam_authorization_clock
FOR EACH ROW EXECUTE FUNCTION iam_guard_authorization_clock_v1();
ALTER TABLE iam_authorization_clock
    ENABLE ALWAYS TRIGGER iam_authorization_clock_step_guard;

CREATE TRIGGER iam_authorization_clock_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_authorization_clock
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_authorization_clock
    ENABLE ALWAYS TRIGGER iam_authorization_clock_removal_guard;

CREATE TRIGGER iam_authorization_operation_receipts_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_authorization_operation_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_authorization_operation_receipts
    ENABLE ALWAYS TRIGGER iam_authorization_operation_receipts_mutation_guard;
