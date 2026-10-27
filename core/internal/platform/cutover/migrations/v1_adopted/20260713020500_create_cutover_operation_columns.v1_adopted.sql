-- +goose Up
CREATE TABLE platform.cutover_operation_columns (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL,
    ordinal smallint NOT NULL,
    source_attnum smallint NOT NULL,
    target_attnum smallint NOT NULL,
    transform_id text NOT NULL,
    primary_key boolean NOT NULL,
    PRIMARY KEY (operation_id, run_id, relation_ordinal, ordinal),
    CONSTRAINT cutover_operation_columns_relation_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal)
        REFERENCES platform.cutover_operation_relations (operation_id, run_id, ordinal)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operation_columns_ordinal_check
        CHECK (ordinal BETWEEN 1 AND 2048),
    CONSTRAINT cutover_operation_columns_attnum_check
        CHECK (
            source_attnum BETWEEN 1 AND 32767 AND
            target_attnum BETWEEN 1 AND 32767
        ),
    CONSTRAINT cutover_operation_columns_transform_check
        CHECK (transform_id ~ '^[a-z][a-z0-9_.-]{0,127}$'),
    CONSTRAINT cutover_operation_columns_source_unique
        UNIQUE (operation_id, run_id, relation_ordinal, source_attnum),
    CONSTRAINT cutover_operation_columns_target_unique
        UNIQUE (operation_id, run_id, relation_ordinal, target_attnum)
);

CREATE INDEX cutover_operation_columns_primary_key_idx
    ON platform.cutover_operation_columns (operation_id, run_id, relation_ordinal)
    WHERE primary_key;
