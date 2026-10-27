-- +goose Up
-- Migration stream version: 20260722140100.
ALTER TABLE iam_one_time_tokens
    DROP CONSTRAINT iam_one_time_tokens_purpose_check,
    ADD CONSTRAINT iam_one_time_tokens_purpose_check CHECK (
        purpose IN (
            'password_reset_v1','email_verification_v1','portal_setup_v1',
            'invitation_accept_v1','initial_admin_setup.v1'
        )
    ),
    DROP CONSTRAINT iam_one_time_tokens_subject_shape_check,
    ADD CONSTRAINT iam_one_time_tokens_subject_shape_check CHECK (
        (
            purpose IN ('password_reset_v1','email_verification_v1',
                'initial_admin_setup.v1')
            AND identifier_kind = 'email_v1'
            AND identifier_revision > 0 AND user_id IS NOT NULL
            AND identifier_id IS NOT NULL AND invitation_id IS NULL
            AND invitation_expires_at IS NULL AND invitation_revision IS NULL
            AND portal_binding_id IS NULL AND portal_setup_revision IS NULL
        ) OR (
            purpose = 'portal_setup_v1' AND identifier_kind = 'portal_v1'
            AND identifier_revision > 0 AND user_id IS NOT NULL
            AND identifier_id IS NOT NULL AND invitation_id IS NULL
            AND invitation_expires_at IS NULL AND invitation_revision IS NULL
            AND portal_binding_id IS NOT NULL AND portal_setup_revision > 0
        ) OR (
            purpose = 'invitation_accept_v1' AND user_id IS NULL
            AND identifier_id IS NULL AND identifier_kind IS NULL
            AND identifier_revision IS NULL AND invitation_id IS NOT NULL
            AND invitation_expires_at IS NOT NULL AND invitation_revision > 0
            AND portal_binding_id IS NULL AND portal_setup_revision IS NULL
        )
    ),
    DROP CONSTRAINT iam_one_time_tokens_ttl_check,
    ADD CONSTRAINT iam_one_time_tokens_ttl_check CHECK (
        (
            purpose IN ('password_reset_v1','portal_setup_v1')
            AND expires_at = created_at + INTERVAL '1 hour'
        ) OR (
            purpose IN ('email_verification_v1','initial_admin_setup.v1')
            AND expires_at = created_at + INTERVAL '24 hours'
        ) OR (
            purpose = 'invitation_accept_v1'
            AND expires_at = invitation_expires_at AND created_at < expires_at
        )
    );

ALTER TABLE iam_lifecycle_delivery_artifacts
    DROP CONSTRAINT iam_lifecycle_delivery_artifacts_purpose_check,
    ADD CONSTRAINT iam_lifecycle_delivery_artifacts_purpose_check CHECK (
        purpose IN ('email_verification_v1','password_reset_v1',
            'portal_setup_v1','initial_admin_setup.v1')
    );

ALTER TABLE iam_lifecycle_delivery_worker_receipts
    DROP CONSTRAINT iam_lifecycle_delivery_worker_receipts_purpose_check,
    ADD CONSTRAINT iam_lifecycle_delivery_worker_receipts_purpose_check CHECK (
        purpose IN ('email_verification_v1','password_reset_v1',
            'portal_setup_v1','initial_admin_setup.v1')
    );

CREATE TABLE iam_initial_admin_setups (
    installation_operation_id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    token_id uuid NOT NULL UNIQUE,
    artifact_id uuid NOT NULL UNIQUE,
    generation bigint NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    accepted_at timestamptz,
    revoked_at timestamptz,
    CONSTRAINT iam_initial_admin_setups_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_initial_admin_setups_token_fk
        FOREIGN KEY (token_id) REFERENCES iam_one_time_tokens (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_initial_admin_setups_shape_check CHECK (
        generation > 0 AND revision > 0
        AND state IN ('pending','accepted','revoked')
        AND (
            state = 'pending' AND accepted_at IS NULL AND revoked_at IS NULL
            OR state = 'accepted' AND accepted_at IS NOT NULL AND revoked_at IS NULL
            OR state = 'revoked' AND accepted_at IS NULL AND revoked_at IS NOT NULL
        )
    )
);
