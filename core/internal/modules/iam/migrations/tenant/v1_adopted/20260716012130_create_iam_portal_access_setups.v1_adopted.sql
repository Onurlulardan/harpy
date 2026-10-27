-- +goose Up
CREATE TABLE iam_portal_access_setups (
    binding_id uuid NOT NULL,
    user_id uuid NOT NULL,
    access_mode_id uuid NOT NULL,
    access_mode text NOT NULL,
    setup_state text NOT NULL,
    approval_state text NOT NULL,
    provenance text NOT NULL,
    approval_revision bigint NOT NULL,
    source_evidence_sha256 bytea NOT NULL,
    revision bigint NOT NULL,
    setup_completed_at timestamptz,
    submitted_at timestamptz,
    decision_at timestamptz,
    decision_actor_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_portal_access_setups_pkey PRIMARY KEY (binding_id),
    CONSTRAINT iam_portal_access_setups_binding_fk
        FOREIGN KEY (binding_id, user_id, access_mode_id)
        REFERENCES iam_portal_subject_bindings (id, user_id, access_mode_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_access_setups_access_mode_fk
        FOREIGN KEY (access_mode_id, user_id, access_mode)
        REFERENCES iam_user_access_modes (id, user_id, mode)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_access_setups_decision_actor_fk
        FOREIGN KEY (decision_actor_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_access_setups_access_mode_check
        CHECK (access_mode = 'portal'),
    CONSTRAINT iam_portal_access_setups_setup_state_check
        CHECK (setup_state IN ('not_required', 'required', 'completed')),
    CONSTRAINT iam_portal_access_setups_approval_state_check CHECK (
        approval_state IN (
            'not_required', 'not_submitted', 'pending',
            'revision_required', 'approved', 'rejected'
        )
    ),
    CONSTRAINT iam_portal_access_setups_provenance_check
        CHECK (provenance IN ('native_v2', 'v1_adopted')),
    CONSTRAINT iam_portal_access_setups_revision_check CHECK (
        revision > 0 AND (
            (approval_state IN ('not_required', 'not_submitted') AND
                approval_revision >= 0) OR
            (approval_state IN (
                'pending', 'revision_required', 'approved', 'rejected'
            ) AND approval_revision > 0)
        )
    ),
    CONSTRAINT iam_portal_access_setups_evidence_hash_check
        CHECK (pg_catalog.octet_length(source_evidence_sha256) = 32),
    CONSTRAINT iam_portal_access_setups_setup_timestamp_check CHECK (
        (setup_state IN ('not_required', 'required') AND
            setup_completed_at IS NULL) OR
        (setup_state = 'completed' AND
            setup_completed_at IS NOT NULL)
    ),
    CONSTRAINT iam_portal_access_setups_state_combination_check CHECK (
        (setup_state = 'required' AND
            approval_state IN ('not_required', 'not_submitted')) OR
        (setup_state IN ('not_required', 'completed') AND
            approval_state IN (
                'not_required', 'not_submitted', 'pending',
                'revision_required', 'approved', 'rejected'
            ))
    ),
    CONSTRAINT iam_portal_access_setups_approval_timestamp_check CHECK (
        (
            approval_state IN ('not_required', 'not_submitted') AND
            submitted_at IS NULL AND decision_at IS NULL AND
            decision_actor_id IS NULL
        ) OR
        (
            approval_state = 'pending' AND
            decision_at IS NULL AND decision_actor_id IS NULL AND
            (submitted_at IS NOT NULL OR provenance = 'v1_adopted')
        ) OR
        (
            approval_state IN (
                'revision_required', 'approved', 'rejected'
            ) AND
            (
                (
                    submitted_at IS NOT NULL AND decision_at IS NOT NULL AND
                    decision_actor_id IS NOT NULL
                ) OR provenance = 'v1_adopted'
            )
        )
    ),
    CONSTRAINT iam_portal_access_setups_timestamp_order_check CHECK (
        updated_at >= created_at AND
        (setup_completed_at IS NULL OR
            (setup_completed_at >= created_at AND
                setup_completed_at <= updated_at)) AND
        (submitted_at IS NULL OR
            (submitted_at >= created_at AND submitted_at <= updated_at)) AND
        (decision_at IS NULL OR
            (decision_at >= created_at AND decision_at <= updated_at)) AND
        (decision_at IS NULL OR submitted_at IS NULL OR
            decision_at >= submitted_at)
    )
);

CREATE INDEX iam_portal_access_setups_state_idx
    ON iam_portal_access_setups (setup_state, approval_state, user_id);
