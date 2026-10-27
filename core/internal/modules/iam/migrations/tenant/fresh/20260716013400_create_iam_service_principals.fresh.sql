-- +goose Up
CREATE TABLE iam_service_principals (
    id uuid NOT NULL,
    principal_key text COLLATE "C" NOT NULL,
    kind text NOT NULL,
    display_name text NOT NULL,
    state text NOT NULL,
    authorization_version bigint NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    disabled_at timestamptz,
    retired_at timestamptz,
    CONSTRAINT iam_service_principals_pkey PRIMARY KEY (id),
    CONSTRAINT iam_service_principals_key UNIQUE (principal_key),
    CONSTRAINT iam_service_principals_key_check CHECK (
        pg_catalog.char_length(principal_key) BETWEEN 3 AND 100 AND
        principal_key ~ '^[a-z][a-z0-9]*([._-][a-z0-9]+)*$'
    ),
    CONSTRAINT iam_service_principals_kind_check CHECK (kind = 'workflow'),
    CONSTRAINT iam_service_principals_display_name_check CHECK (
        pg_catalog.char_length(display_name) BETWEEN 1 AND 100 AND
        display_name = pg_catalog.btrim(display_name)
    ),
    CONSTRAINT iam_service_principals_state_check
        CHECK (state IN ('active', 'disabled', 'retired')),
    CONSTRAINT iam_service_principals_authorization_version_check
        CHECK (authorization_version > 0),
    CONSTRAINT iam_service_principals_revision_check CHECK (revision > 0),
    CONSTRAINT iam_service_principals_timestamp_order_check CHECK (
        updated_at >= created_at AND
        (disabled_at IS NULL OR
            (disabled_at >= created_at AND disabled_at <= updated_at)) AND
        (retired_at IS NULL OR
            (retired_at >= created_at AND retired_at <= updated_at))
    ),
    CONSTRAINT iam_service_principals_lifecycle_check CHECK (
        (state = 'active' AND disabled_at IS NULL AND retired_at IS NULL) OR
        (state = 'disabled' AND disabled_at IS NOT NULL AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND
            (disabled_at IS NULL OR retired_at >= disabled_at))
    )
);

CREATE INDEX iam_service_principals_state_idx
    ON iam_service_principals (state, kind, id);

-- +goose StatementBegin
CREATE FUNCTION iam_guard_service_principal_terminal_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR OLD.state IS DISTINCT FROM 'retired' THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING
        ERRCODE = '23514',
        CONSTRAINT = 'iam_service_principals_terminal_guard',
        MESSAGE = 'Retired IAM service principal is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_service_principal_terminal_v1() FROM PUBLIC;

CREATE TRIGGER iam_service_principals_revision_guard
BEFORE UPDATE ON iam_service_principals
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_service_principals
    ENABLE ALWAYS TRIGGER iam_service_principals_revision_guard;

CREATE TRIGGER iam_service_principals_immutable_guard
BEFORE UPDATE ON iam_service_principals
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'principal_key', 'kind', 'created_at'
);
ALTER TABLE iam_service_principals
    ENABLE ALWAYS TRIGGER iam_service_principals_immutable_guard;

CREATE TRIGGER iam_service_principals_state_guard
BEFORE UPDATE ON iam_service_principals
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'active>disabled', 'disabled>active',
    'active>retired', 'disabled>retired'
);
ALTER TABLE iam_service_principals
    ENABLE ALWAYS TRIGGER iam_service_principals_state_guard;

CREATE TRIGGER iam_service_principals_terminal_guard
BEFORE UPDATE ON iam_service_principals
FOR EACH ROW EXECUTE FUNCTION iam_guard_service_principal_terminal_v1();
ALTER TABLE iam_service_principals
    ENABLE ALWAYS TRIGGER iam_service_principals_terminal_guard;

CREATE TRIGGER iam_service_principals_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_service_principals
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_service_principals
    ENABLE ALWAYS TRIGGER iam_service_principals_removal_guard;
