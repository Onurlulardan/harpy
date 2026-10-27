-- +goose Up
CREATE TABLE iam_mcp_oauth_grants (
    id uuid PRIMARY KEY,
    authorization_request_id uuid NOT NULL UNIQUE,
    user_id uuid NOT NULL,
    session_id uuid NOT NULL,
    client_id text COLLATE "C" NOT NULL,
    resource_uri text NOT NULL,
    scope text COLLATE "C" NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    platform_authorization_clock bigint NOT NULL,
    security_epoch bytea NOT NULL,
    auth_time timestamptz NOT NULL,
    acr text COLLATE "C" NOT NULL,
    amr text[] NOT NULL,
    state text COLLATE "C" NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_mcp_oauth_grants_request_fk
        FOREIGN KEY (authorization_request_id)
        REFERENCES platform.iam_mcp_authorization_requests (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_grants_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_grants_session_fk
        FOREIGN KEY (session_id, user_id) REFERENCES iam_sessions (id, user_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_grants_contract_ck CHECK (
        pg_catalog.char_length(client_id) BETWEEN 1 AND 128 AND
        pg_catalog.char_length(resource_uri) BETWEEN 1 AND 2048 AND
        scope = 'mcp:use' AND token_version > 0 AND authorization_version > 0 AND
        platform_authorization_clock >= 0 AND
        pg_catalog.octet_length(security_epoch) = 32 AND
        acr IN ('urn:harpy:acr:baseline','urn:harpy:acr:step-up') AND
        pg_catalog.cardinality(amr) BETWEEN 1 AND 3 AND
        amr <@ ARRAY['pwd','webauthn','totp']::text[] AND revision > 0
    ),
    CONSTRAINT iam_mcp_oauth_grants_state_ck CHECK (
        state IN ('active','revoked','expired') AND
        ((state = 'active' AND revoked_at IS NULL) OR
         (state = 'revoked' AND revoked_at IS NOT NULL) OR
         (state = 'expired' AND revoked_at IS NULL))
    ),
    CONSTRAINT iam_mcp_oauth_grants_time_ck CHECK (
        auth_time <= created_at AND updated_at >= created_at AND
        absolute_expires_at = created_at + INTERVAL '90 days' AND
        (revoked_at IS NULL OR revoked_at BETWEEN created_at AND updated_at) AND
        retain_until >= absolute_expires_at + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_mcp_oauth_grants_session_idx
    ON iam_mcp_oauth_grants (session_id, state);
CREATE INDEX iam_mcp_oauth_grants_retention_idx
    ON iam_mcp_oauth_grants (retain_until);

CREATE TABLE iam_mcp_oauth_access_tokens (
    jti uuid PRIMARY KEY,
    grant_id uuid NOT NULL,
    token_sha256 bytea NOT NULL UNIQUE,
    state text COLLATE "C" NOT NULL,
    revision bigint NOT NULL,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_mcp_oauth_access_tokens_grant_fk
        FOREIGN KEY (grant_id) REFERENCES iam_mcp_oauth_grants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_access_tokens_contract_ck CHECK (
        pg_catalog.octet_length(token_sha256) = 32 AND revision > 0 AND
        state IN ('active','revoked','expired') AND
        ((state = 'active' AND revoked_at IS NULL) OR
         (state = 'revoked' AND revoked_at IS NOT NULL) OR
         (state = 'expired' AND revoked_at IS NULL))
    ),
    CONSTRAINT iam_mcp_oauth_access_tokens_time_ck CHECK (
        expires_at = issued_at + INTERVAL '15 minutes' AND
        (revoked_at IS NULL OR revoked_at BETWEEN issued_at AND expires_at) AND
        retain_until >= expires_at + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_mcp_oauth_access_tokens_grant_idx
    ON iam_mcp_oauth_access_tokens (grant_id, state, expires_at);
CREATE INDEX iam_mcp_oauth_access_tokens_retention_idx
    ON iam_mcp_oauth_access_tokens (retain_until);
