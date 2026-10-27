-- +goose Up
CREATE TABLE iam_permission_request_groups (
    id uuid NOT NULL,
    requester_user_id uuid NOT NULL,
    request_mode text NOT NULL,
    creator_authority_permission_id uuid NOT NULL,
    reviewer_authority_permission_id uuid NOT NULL,
    step_up_purpose_sha256 bytea NOT NULL,
    state text NOT NULL,
    request_count smallint NOT NULL,
    reviewer_user_id uuid,
    review_task_id uuid,
    revision bigint NOT NULL,
    expires_at timestamptz NOT NULL,
    terminal_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_permission_request_groups_pkey PRIMARY KEY (id),
    CONSTRAINT iam_permission_request_groups_state_identity_key
        UNIQUE (id, state),
    CONSTRAINT iam_permission_request_groups_requester_fk
        FOREIGN KEY (requester_user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_groups_creator_permission_fk
        FOREIGN KEY (creator_authority_permission_id) REFERENCES iam_permissions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_groups_reviewer_permission_fk
        FOREIGN KEY (reviewer_authority_permission_id) REFERENCES iam_permissions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_groups_reviewer_fk
        FOREIGN KEY (reviewer_user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_groups_mode_check
        CHECK (request_mode IN ('SELF_SERVICE', 'PRIVILEGED')),
    CONSTRAINT iam_permission_request_groups_state_check CHECK (
        state IN (
            'PENDING', 'IN_REVIEW', 'APPROVED', 'REJECTED',
            'CANCELLED', 'EXPIRED'
        )
    ),
    CONSTRAINT iam_permission_request_groups_count_check
        CHECK (request_count BETWEEN 1 AND 100),
    CONSTRAINT iam_permission_request_groups_revision_check CHECK (revision > 0),
    CONSTRAINT iam_permission_request_groups_step_up_check
        CHECK (pg_catalog.octet_length(step_up_purpose_sha256) = 32),
    CONSTRAINT iam_permission_request_groups_timestamp_check CHECK (
        updated_at >= created_at AND expires_at > created_at AND
        (terminal_at IS NULL OR
            (terminal_at >= created_at AND terminal_at <= updated_at))
    ),
    CONSTRAINT iam_permission_request_groups_lifecycle_check CHECK (
        (state = 'PENDING' AND reviewer_user_id IS NULL AND
            review_task_id IS NULL AND terminal_at IS NULL) OR
        (state = 'IN_REVIEW' AND reviewer_user_id IS NOT NULL AND
            review_task_id IS NOT NULL AND terminal_at IS NULL) OR
        (state IN ('APPROVED', 'REJECTED') AND reviewer_user_id IS NOT NULL AND
            review_task_id IS NOT NULL AND terminal_at IS NOT NULL) OR
        (state IN ('CANCELLED', 'EXPIRED') AND terminal_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX iam_permission_request_groups_review_task_uidx
    ON iam_permission_request_groups (review_task_id)
    WHERE review_task_id IS NOT NULL;

CREATE INDEX iam_permission_request_groups_requester_state_idx
    ON iam_permission_request_groups (requester_user_id, state, created_at, id);

CREATE INDEX iam_permission_request_groups_reviewer_state_idx
    ON iam_permission_request_groups (reviewer_user_id, state, updated_at, id)
    WHERE reviewer_user_id IS NOT NULL;
