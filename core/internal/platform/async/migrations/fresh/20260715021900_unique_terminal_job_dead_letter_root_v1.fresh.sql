-- +goose Up
CREATE UNIQUE INDEX dead_letters_terminal_job_root_unique
ON platform.dead_letters (source_id)
WHERE record_kind = 'root' AND source_kind = 'job'
  AND reason_class IN ('permanent','delivery_limit','outcome_unknown');
