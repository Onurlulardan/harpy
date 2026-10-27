-- +goose Up
CREATE TABLE platform.iam_post_restore_reconciliations (
    restore_operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    recovery_set_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    anchor_id uuid NOT NULL,
    new_database_system_id uuid NOT NULL,
    security_epoch bytea NOT NULL CHECK (pg_catalog.octet_length(security_epoch)=32),
    restore_receipt_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(restore_receipt_sha256)=32),
    protected_head_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(protected_head_sha256)=32),
    custody_generation bigint NOT NULL CHECK (custody_generation>0),
    custody_result_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(custody_result_sha256)=32),
    receipt jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(receipt)='object'),
    state text NOT NULL CHECK (state IN (
        'identity_reactivation_required','reactivated','fenced'
    )),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at>=created_at),
    reactivated_at timestamptz,
    CONSTRAINT iam_post_restore_environment_uq UNIQUE(environment_id),
    CONSTRAINT iam_post_restore_anchor_fk FOREIGN KEY(anchor_id)
        REFERENCES platform.iam_recovery_anchors(anchor_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_post_restore_state_ck CHECK (
        (state='identity_reactivation_required' AND reactivated_at IS NULL) OR
        (state='reactivated' AND reactivated_at IS NOT NULL) OR state='fenced'
    )
);

CREATE TABLE platform.iam_post_restore_challenges (
    operation_id uuid PRIMARY KEY,
    restore_operation_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    anchor_id uuid NOT NULL,
    challenge text COLLATE "C" NOT NULL CHECK (challenge ~ '^[A-Za-z0-9_-]{43}$'),
    context_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(context_sha256)=32),
    rp_id text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(rp_id) BETWEEN 1 AND 253),
    origin text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(origin) BETWEEN 1 AND 512),
    state text NOT NULL CHECK (state IN ('pending','consumed','expired','failed')),
    attempt_count smallint NOT NULL DEFAULT 0 CHECK (attempt_count BETWEEN 0 AND 5),
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL CHECK (
        expires_at>created_at AND expires_at<=created_at+INTERVAL '5 minutes'
    ),
    terminal_at timestamptz,
    CONSTRAINT iam_post_restore_challenge_reconcile_fk FOREIGN KEY(restore_operation_id)
        REFERENCES platform.iam_post_restore_reconciliations(restore_operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_post_restore_challenge_state_ck CHECK (
        (state='pending' AND terminal_at IS NULL) OR
        (state<>'pending' AND terminal_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX iam_post_restore_one_pending_challenge_uq
ON platform.iam_post_restore_challenges(restore_operation_id)
WHERE state='pending';

CREATE TABLE platform.iam_post_restore_reactivation_receipts (
    restore_operation_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL UNIQUE,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    anchor_id uuid NOT NULL,
    assertion_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(assertion_sha256)=32),
    password_credential_sha256 bytea NOT NULL CHECK (
        pg_catalog.octet_length(password_credential_sha256)=32
    ),
    receipt_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(receipt_sha256)=32),
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_post_restore_reactivation_reconcile_fk FOREIGN KEY(restore_operation_id)
        REFERENCES platform.iam_post_restore_reconciliations(restore_operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);
