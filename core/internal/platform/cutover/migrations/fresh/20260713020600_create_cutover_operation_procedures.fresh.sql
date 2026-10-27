-- +goose Up
CREATE TABLE platform.cutover_operation_procedures (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    ordinal smallint NOT NULL,
    relation_ordinal smallint,
    procedure_kind text NOT NULL,
    procedure_oid bigint NOT NULL,
    procedure_schema text NOT NULL,
    procedure_name text NOT NULL,
    identity_arguments text NOT NULL,
    procedure_signature text NOT NULL,
    PRIMARY KEY (operation_id, run_id, ordinal),
    CONSTRAINT cutover_operation_procedures_operation_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.cutover_operations (operation_id, run_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operation_procedures_relation_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal)
        REFERENCES platform.cutover_operation_relations (operation_id, run_id, ordinal)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operation_procedures_ordinal_check
        CHECK (ordinal BETWEEN 1 AND 12288),
    CONSTRAINT cutover_operation_procedures_kind_check
        CHECK (procedure_kind IN ('apply', 'checkpoint', 'receipt')),
    CONSTRAINT cutover_operation_procedures_binding_check
        CHECK (
            (procedure_kind = 'apply' AND relation_ordinal IS NOT NULL) OR
            (procedure_kind IN ('checkpoint', 'receipt') AND relation_ordinal IS NULL)
        ),
    CONSTRAINT cutover_operation_procedures_oid_check
        CHECK (procedure_oid BETWEEN 1 AND 4294967295),
    CONSTRAINT cutover_operation_procedures_name_check
        CHECK (
            procedure_schema ~ '^[a-z][a-z0-9_]{0,62}$' AND
            procedure_name ~ '^[a-z][a-z0-9_]{0,62}$'
        ),
    CONSTRAINT cutover_operation_procedures_arguments_check
        CHECK (pg_catalog.octet_length(identity_arguments) BETWEEN 1 AND 1024),
    CONSTRAINT cutover_operation_procedures_signature_check
        CHECK (pg_catalog.octet_length(procedure_signature) BETWEEN 1 AND 1024),
    CONSTRAINT cutover_operation_procedures_oid_unique
        UNIQUE (operation_id, run_id, procedure_oid)
);

CREATE UNIQUE INDEX cutover_operation_procedures_binding_idx
    ON platform.cutover_operation_procedures (
        operation_id,
        run_id,
        procedure_kind,
        COALESCE(relation_ordinal, 0::smallint)
    );
