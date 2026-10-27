-- +goose Up
CREATE TABLE platform.cutover_operation_checkpoint_keys (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL,
    ordinal integer NOT NULL,
    checkpoint_key text NOT NULL,
    PRIMARY KEY (operation_id, run_id, relation_ordinal, checkpoint_key),
    CONSTRAINT cutover_operation_checkpoint_keys_relation_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal)
        REFERENCES platform.cutover_operation_relations (operation_id, run_id, ordinal)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operation_checkpoint_keys_ordinal_check
        CHECK (ordinal BETWEEN 1 AND 1024),
    CONSTRAINT cutover_operation_checkpoint_keys_key_check
        CHECK (
            pg_catalog.octet_length(checkpoint_key) BETWEEN 1 AND 512 AND
            checkpoint_key !~ '[[:cntrl:]]'
        ),
    CONSTRAINT cutover_operation_checkpoint_keys_ordinal_unique
        UNIQUE (operation_id, run_id, relation_ordinal, ordinal)
);
