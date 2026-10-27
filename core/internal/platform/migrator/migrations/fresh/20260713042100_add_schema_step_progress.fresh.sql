-- +goose Up
ALTER TABLE platform.schema_job_steps
    ADD COLUMN checkpoint_record_id uuid,
    ADD COLUMN processed_rows bigint NOT NULL DEFAULT 0,
    ADD CONSTRAINT schema_job_steps_progress_check CHECK (
        (operation_id = 'backfill_field_batch' AND (
            (processed_rows = 0 AND checkpoint_record_id IS NULL) OR
            (processed_rows > 0 AND checkpoint_record_id IS NOT NULL)
        )) OR
        (operation_id <> 'backfill_field_batch' AND
            processed_rows = 0 AND checkpoint_record_id IS NULL)
    );
