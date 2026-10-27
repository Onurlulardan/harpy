-- +goose Up
ALTER TABLE platform.cutover_operation_procedures
    ADD COLUMN procedure_contract_id text NOT NULL
        DEFAULT 'platform.legacy_unregistered.v1';

ALTER TABLE platform.cutover_operation_procedures
    ALTER COLUMN procedure_contract_id DROP DEFAULT,
    ADD CONSTRAINT cutover_operation_procedures_contract_id_check
        CHECK (procedure_contract_id ~ '^[a-z][a-z0-9_.-]{0,127}$'),
    ADD CONSTRAINT cutover_operation_procedures_contract_id_unique
        UNIQUE (operation_id, run_id, procedure_contract_id);
