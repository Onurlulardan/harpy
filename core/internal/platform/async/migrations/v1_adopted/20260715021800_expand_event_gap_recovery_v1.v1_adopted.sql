-- +goose Up
ALTER TABLE platform.dead_letters
    DROP CONSTRAINT dead_letters_reason_class_check,
    ADD CONSTRAINT dead_letters_reason_class_check CHECK (reason_class IN (
        'permanent', 'unknown_version', 'delivery_limit',
        'outcome_unknown', 'stale_generation', 'sequence_gap'
    )),
    DROP CONSTRAINT dead_letters_payload_canonical_check,
    ADD CONSTRAINT dead_letters_payload_canonical_check CHECK (
        payload_canonical IS NULL OR
        pg_catalog.octet_length(payload_canonical) BETWEEN 2 AND 131072
    );

