-- +goose Up
CREATE TABLE work_id_tombstones (
    target_kind text NOT NULL PRIMARY KEY,
    target_id uuid NOT NULL,
    reference_epoch text NOT NULL,
    receipt_sha256 text NOT NULL,
    purged_at timestamptz,
    retain_until text NOT NULL
);

-- +goose Down
DROP TABLE work_id_tombstones;
