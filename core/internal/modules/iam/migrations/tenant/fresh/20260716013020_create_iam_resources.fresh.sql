-- +goose Up
CREATE TABLE iam_resources (
    id uuid NOT NULL,
    resource_key text COLLATE "C" NOT NULL,
    owner_domain text COLLATE "C" NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    display_name text NOT NULL,
    description text,
    is_protected boolean NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_resources_pkey PRIMARY KEY (id),
    CONSTRAINT iam_resources_key_key UNIQUE (resource_key),
    CONSTRAINT iam_resources_key_check CHECK (
        pg_catalog.char_length(resource_key) BETWEEN 1 AND 100 AND
        resource_key ~ '^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$'
    ),
    CONSTRAINT iam_resources_owner_domain_check CHECK (
        pg_catalog.char_length(owner_domain) BETWEEN 1 AND 100 AND
        owner_domain ~ '^[a-z][a-z0-9_]*$'
    ),
    CONSTRAINT iam_resources_kind_check CHECK (
        pg_catalog.char_length(resource_kind) BETWEEN 1 AND 100 AND
        resource_kind ~ '^[a-z][a-z0-9_]*$'
    ),
    CONSTRAINT iam_resources_display_name_check CHECK (
        pg_catalog.char_length(display_name) BETWEEN 1 AND 255 AND
        display_name = pg_catalog.btrim(display_name)
    ),
    CONSTRAINT iam_resources_description_check CHECK (
        description IS NULL OR pg_catalog.char_length(description) <= 2000
    ),
    CONSTRAINT iam_resources_state_check CHECK (state IN ('active', 'retired')),
    CONSTRAINT iam_resources_revision_check CHECK (revision > 0),
    CONSTRAINT iam_resources_timestamp_order_check CHECK (updated_at >= created_at),
    CONSTRAINT iam_resources_retirement_check CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

-- +goose StatementBegin
CREATE FUNCTION iam_forbid_rbac_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '23514',
        CONSTRAINT = TG_TABLE_NAME || '_forbidden_mutation_guard',
        MESSAGE = 'IAM RBAC mutation requires an explicit lifecycle transition';
END;
$function$;
REVOKE ALL ON FUNCTION iam_forbid_rbac_mutation_v1() FROM PUBLIC;

CREATE FUNCTION iam_guard_rbac_terminal_v1()
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
        CONSTRAINT = TG_TABLE_NAME || '_terminal_guard',
        MESSAGE = 'Retired IAM RBAC state is immutable';
END;
$function$;
REVOKE ALL ON FUNCTION iam_guard_rbac_terminal_v1() FROM PUBLIC;
-- +goose StatementEnd

CREATE TRIGGER iam_resources_revision_guard
BEFORE UPDATE ON iam_resources
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_resources ENABLE ALWAYS TRIGGER iam_resources_revision_guard;

CREATE TRIGGER iam_resources_immutable_guard
BEFORE UPDATE ON iam_resources
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'resource_key', 'owner_domain', 'resource_kind',
    'is_protected', 'created_at'
);
ALTER TABLE iam_resources ENABLE ALWAYS TRIGGER iam_resources_immutable_guard;

CREATE TRIGGER iam_resources_state_guard
BEFORE UPDATE ON iam_resources
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>retired');
ALTER TABLE iam_resources ENABLE ALWAYS TRIGGER iam_resources_state_guard;

CREATE TRIGGER iam_resources_terminal_guard
BEFORE UPDATE ON iam_resources
FOR EACH ROW EXECUTE FUNCTION iam_guard_rbac_terminal_v1();
ALTER TABLE iam_resources ENABLE ALWAYS TRIGGER iam_resources_terminal_guard;

CREATE TRIGGER iam_resources_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_resources
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_resources ENABLE ALWAYS TRIGGER iam_resources_removal_guard;
