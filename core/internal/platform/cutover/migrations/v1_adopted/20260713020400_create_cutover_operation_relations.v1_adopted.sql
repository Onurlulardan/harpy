-- +goose Up
CREATE TABLE platform.cutover_operation_relations (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    ordinal smallint NOT NULL,
    source_schema_oid bigint NOT NULL,
    source_relation_oid bigint NOT NULL,
    source_schema_name text NOT NULL,
    source_relation_name text NOT NULL,
    target_schema_oid bigint NOT NULL,
    target_relation_oid bigint NOT NULL,
    target_schema_name text NOT NULL,
    target_relation_name text NOT NULL,
    replica_identity text NOT NULL,
    PRIMARY KEY (operation_id, run_id, ordinal),
    CONSTRAINT cutover_operation_relations_operation_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.cutover_operations (operation_id, run_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operation_relations_ordinal_check
        CHECK (ordinal BETWEEN 1 AND 4096),
    CONSTRAINT cutover_operation_relations_source_oid_check
        CHECK (
            source_schema_oid BETWEEN 1 AND 4294967295 AND
            source_relation_oid BETWEEN 1 AND 4294967295
        ),
    CONSTRAINT cutover_operation_relations_target_oid_check
        CHECK (
            target_schema_oid BETWEEN 1 AND 4294967295 AND
            target_relation_oid BETWEEN 1 AND 4294967295
        ),
    CONSTRAINT cutover_operation_relations_source_name_check
        CHECK (
            source_schema_name ~ '^[a-z][a-z0-9_]{0,62}$' AND
            source_relation_name ~ '^[a-z][a-z0-9_]{0,62}$'
        ),
    CONSTRAINT cutover_operation_relations_target_name_check
        CHECK (
            target_schema_name ~ '^[a-z][a-z0-9_]{0,62}$' AND
            target_relation_name ~ '^[a-z][a-z0-9_]{0,62}$'
        ),
    CONSTRAINT cutover_operation_relations_identity_check
        CHECK (replica_identity IN ('primary_key', 'index')),
    CONSTRAINT cutover_operation_relations_source_unique
        UNIQUE (operation_id, run_id, source_schema_oid, source_relation_oid),
    CONSTRAINT cutover_operation_relations_target_unique
        UNIQUE (operation_id, run_id, target_schema_oid, target_relation_oid)
);
