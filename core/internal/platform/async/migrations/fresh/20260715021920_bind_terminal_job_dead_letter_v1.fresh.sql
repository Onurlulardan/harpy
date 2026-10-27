-- +goose Up
CREATE TRIGGER jobs_terminal_dead_letter_guard
AFTER UPDATE OF state ON platform.jobs
FOR EACH ROW
WHEN (OLD.state IS DISTINCT FROM 'dead' AND NEW.state = 'dead')
EXECUTE FUNCTION platform.platform_async_record_terminal_job_dead_letter_v1();

ALTER TABLE platform.jobs
    ENABLE ALWAYS TRIGGER jobs_terminal_dead_letter_guard;
