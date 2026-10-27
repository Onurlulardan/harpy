-- +goose Up
CREATE TABLE work_boards (
    id uuid PRIMARY KEY,
    creator_user_id uuid NOT NULL,
    name text NOT NULL CHECK (pg_catalog.octet_length(name) BETWEEN 1 AND 800),
    description text NOT NULL CHECK (pg_catalog.octet_length(description) <= 64000),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('ACTIVE','ARCHIVED')),
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    archived_at timestamptz,
    CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (creator_user_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    CHECK (pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)),
    CHECK ((lifecycle_state = 'ACTIVE' AND archived_at IS NULL) OR
           (lifecycle_state = 'ARCHIVED' AND archived_at IS NOT NULL))
);

CREATE TABLE work_board_columns (
    id uuid PRIMARY KEY,
    board_id uuid NOT NULL,
    name text NOT NULL CHECK (pg_catalog.octet_length(name) BETWEEN 1 AND 800),
    status_category text NOT NULL CHECK (status_category IN (
        'OPEN','IN_PROGRESS','IN_REVIEW','DONE'
    )),
    position bigint NOT NULL CHECK (position > 0),
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT work_board_columns_board_fk FOREIGN KEY (board_id)
        REFERENCES work_boards(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT work_board_columns_position_key UNIQUE (board_id, position)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE work_board_memberships (
    id uuid PRIMARY KEY,
    board_id uuid NOT NULL,
    principal_kind text NOT NULL CHECK (principal_kind IN ('USER','ROLE','ORGANIZATION')),
    principal_id uuid NOT NULL,
    board_role text NOT NULL CHECK (board_role IN ('VIEWER','EDITOR','ADMIN')),
    active boolean NOT NULL,
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deactivated_at timestamptz,
    CONSTRAINT work_board_memberships_board_fk FOREIGN KEY (board_id)
        REFERENCES work_boards(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK ((active AND deactivated_at IS NULL) OR (NOT active AND deactivated_at IS NOT NULL))
);

CREATE UNIQUE INDEX work_board_memberships_active_key
    ON work_board_memberships(board_id, principal_kind, principal_id) WHERE active;
CREATE INDEX work_board_columns_list_idx ON work_board_columns(board_id, position, id);
CREATE INDEX work_board_memberships_principal_idx
    ON work_board_memberships(principal_kind, principal_id, board_id) WHERE active;
