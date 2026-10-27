-- +goose Up
CREATE TABLE platform.iam_mcp_authorization_requests (
    id uuid PRIMARY KEY,
    operation_id text COLLATE "C" NOT NULL UNIQUE,
    request_sha256 bytea NOT NULL,
    client_id text COLLATE "C" NOT NULL,
    redirect_uri text NOT NULL,
    resource_uri text NOT NULL,
    scope text COLLATE "C" NOT NULL,
    state_value text NOT NULL,
    code_challenge text COLLATE "C" NOT NULL,
    code_challenge_method text COLLATE "C" NOT NULL,
    status text COLLATE "C" NOT NULL,
    revision bigint NOT NULL,
    tenant_id uuid,
    user_id uuid,
    session_id uuid,
    grant_id uuid,
    decision_operation_id text COLLATE "C",
    decision_sha256 bytea,
    code_selector bytea UNIQUE,
    code_hash_key_id text COLLATE "C",
    code_secret_hmac bytea,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    decided_at timestamptz,
    code_expires_at timestamptz,
    consumed_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_mcp_authorization_requests_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_authorization_requests_request_ck CHECK (
        pg_catalog.char_length(operation_id) BETWEEN 1 AND 128 AND
        pg_catalog.octet_length(request_sha256) = 32 AND
        pg_catalog.char_length(client_id) BETWEEN 1 AND 128 AND
        pg_catalog.char_length(redirect_uri) BETWEEN 1 AND 2048 AND
        pg_catalog.char_length(resource_uri) BETWEEN 1 AND 2048 AND
        scope = 'mcp:use' AND
        pg_catalog.char_length(state_value) BETWEEN 16 AND 512 AND
        pg_catalog.char_length(code_challenge) = 43 AND
        code_challenge_method = 'S256'
    ),
    CONSTRAINT iam_mcp_authorization_requests_state_ck CHECK (
        status IN ('pending','approved','denied','consumed','revoked') AND
        revision > 0
    ),
    CONSTRAINT iam_mcp_authorization_requests_decision_ck CHECK (
        (status = 'pending' AND tenant_id IS NULL AND user_id IS NULL AND
            session_id IS NULL AND grant_id IS NULL AND
            decision_operation_id IS NULL AND decision_sha256 IS NULL AND
            code_selector IS NULL AND code_hash_key_id IS NULL AND
            code_secret_hmac IS NULL AND decided_at IS NULL AND
            code_expires_at IS NULL AND consumed_at IS NULL) OR
        (status = 'denied' AND tenant_id IS NOT NULL AND user_id IS NOT NULL AND
            session_id IS NOT NULL AND grant_id IS NULL AND
            decision_operation_id IS NOT NULL AND decision_sha256 IS NOT NULL AND
            code_selector IS NULL AND code_hash_key_id IS NULL AND
            code_secret_hmac IS NULL AND decided_at IS NOT NULL AND
            code_expires_at IS NULL AND consumed_at IS NULL) OR
        (status IN ('approved','consumed','revoked') AND
            tenant_id IS NOT NULL AND user_id IS NOT NULL AND
            session_id IS NOT NULL AND grant_id IS NOT NULL AND
            decision_operation_id IS NOT NULL AND decision_sha256 IS NOT NULL AND
            code_selector IS NOT NULL AND code_hash_key_id IS NOT NULL AND
            code_secret_hmac IS NOT NULL AND decided_at IS NOT NULL AND
            code_expires_at IS NOT NULL AND
            ((status = 'consumed' AND consumed_at IS NOT NULL) OR
             (status <> 'consumed' AND consumed_at IS NULL)))
    ),
    CONSTRAINT iam_mcp_authorization_requests_secret_ck CHECK (
        (code_selector IS NULL AND code_hash_key_id IS NULL AND
            code_secret_hmac IS NULL) OR
        (pg_catalog.octet_length(code_selector) = 16 AND
            pg_catalog.char_length(code_hash_key_id) BETWEEN 1 AND 128 AND
            pg_catalog.octet_length(code_secret_hmac) = 32)
    ),
    CONSTRAINT iam_mcp_authorization_requests_time_ck CHECK (
        expires_at = created_at + INTERVAL '10 minutes' AND
        (decided_at IS NULL OR decided_at BETWEEN created_at AND expires_at) AND
        (code_expires_at IS NULL OR
            code_expires_at BETWEEN decided_at AND decided_at + INTERVAL '5 minutes') AND
        (consumed_at IS NULL OR consumed_at BETWEEN decided_at AND code_expires_at) AND
        retain_until >= expires_at + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_mcp_authorization_requests_retention_idx
    ON platform.iam_mcp_authorization_requests (retain_until);

CREATE TABLE platform.iam_mcp_credential_locators (
    selector bytea PRIMARY KEY,
    credential_kind text COLLATE "C" NOT NULL,
    tenant_id uuid NOT NULL,
    request_id uuid,
    grant_id uuid NOT NULL,
    family_id uuid,
    token_id uuid,
    source_revision bigint NOT NULL,
    source_sha256 bytea NOT NULL,
    state text COLLATE "C" NOT NULL,
    created_at timestamptz NOT NULL,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_mcp_credential_locators_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_credential_locators_request_fk
        FOREIGN KEY (request_id)
        REFERENCES platform.iam_mcp_authorization_requests (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_credential_locators_shape_ck CHECK (
        pg_catalog.octet_length(selector) = 16 AND
        credential_kind IN ('authorization_code','refresh_token') AND
        source_revision > 0 AND pg_catalog.octet_length(source_sha256) = 32 AND
        state IN ('active','consumed','revoked','expired') AND
        ((credential_kind = 'authorization_code' AND request_id IS NOT NULL AND
            family_id IS NULL AND token_id IS NULL) OR
         (credential_kind = 'refresh_token' AND request_id IS NULL AND
            family_id IS NOT NULL AND token_id IS NOT NULL))
    ),
    CONSTRAINT iam_mcp_credential_locators_time_ck
        CHECK (retain_until > created_at)
);

CREATE INDEX iam_mcp_credential_locators_retention_idx
    ON platform.iam_mcp_credential_locators (retain_until);
