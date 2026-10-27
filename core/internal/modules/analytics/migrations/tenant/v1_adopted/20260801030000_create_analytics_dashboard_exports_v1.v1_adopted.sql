-- +goose Up
CREATE TABLE analytics_dashboard_exports (
    export_id uuid PRIMARY KEY,
    dashboard_id uuid NOT NULL,
    dashboard_version_id uuid NOT NULL,
    job_id uuid NOT NULL UNIQUE,
    actor_id uuid NOT NULL,
    format text NOT NULL,
    state text NOT NULL,
    row_revision bigint NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    dashboard_manifest_sha256 bytea NOT NULL,
    pin_sha256 bytea NOT NULL,
    parameters_document bytea NOT NULL,
    parameters_sha256 bytea NOT NULL,
    pinned_widgets jsonb NOT NULL,
    widget_count smallint NOT NULL,
    maximum_rows bigint NOT NULL,
    maximum_plaintext_bytes bigint NOT NULL,
    attempt_generation bigint NOT NULL DEFAULT 0,
    lease_owner uuid,
    lease_until timestamptz,
    cancellation_requested_at timestamptz,
    effect_started_at timestamptz,
    object_key text,
    object_version text,
    key_id text,
    wrapped_dek bytea,
    wrap_nonce bytea,
    header_sha256 bytea,
    ciphertext_sha256 bytea,
    plaintext_sha256 bytea,
    ciphertext_bytes bigint,
    plaintext_bytes bigint,
    row_count bigint,
    filename text,
    content_type text,
    provider_request_id text,
    failure_code text,
    delete_operation_id uuid,
    delete_job_id uuid,
    delete_actor_id uuid,
    delete_idempotency_key text,
    delete_request_sha256 bytea,
	delete_attempt_generation bigint NOT NULL DEFAULT 0,
	delete_lease_owner uuid,
	delete_lease_until timestamptz,
	delete_provider_request_id text,
	delete_response_sha256 bytea,
	delete_absence_sha256 bytea,
	delete_verified_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    ready_at timestamptz,
    expires_at timestamptz NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT analytics_dashboard_exports_dashboard_fk FOREIGN KEY (
        dashboard_id
    ) REFERENCES analytics_dashboards(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_exports_version_fk FOREIGN KEY (
        dashboard_version_id,dashboard_id
    ) REFERENCES analytics_dashboard_versions(id,dashboard_id)
      ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_exports_idempotency_key UNIQUE (
        actor_id,idempotency_key
    ),
    CONSTRAINT analytics_dashboard_exports_identity_check CHECK (
        pg_catalog.uuid_extract_version(export_id)=7
        AND pg_catalog.uuid_extract_version(job_id)=7
        AND actor_id<>'00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.octet_length(idempotency_key) BETWEEN 16 AND 128
        AND pg_catalog.octet_length(request_sha256)=32
    ),
    CONSTRAINT analytics_dashboard_exports_pin_check CHECK (
        format IN ('csv','xlsx','pdf')
        AND pg_catalog.octet_length(dashboard_manifest_sha256)=32
        AND pg_catalog.octet_length(pin_sha256)=32
        AND pg_catalog.octet_length(parameters_document) BETWEEN 2 AND 65536
        AND pg_catalog.sha256(parameters_document)=parameters_sha256
        AND pg_catalog.pg_input_is_valid(
            pg_catalog.convert_from(parameters_document,'UTF8'),'jsonb')
        AND pg_catalog.jsonb_typeof(pinned_widgets)='array'
        AND widget_count=pg_catalog.jsonb_array_length(pinned_widgets)
        AND widget_count BETWEEN 1 AND 30
        AND maximum_rows BETWEEN 1 AND 100000
        AND maximum_plaintext_bytes BETWEEN 1 AND 67108864
        AND token_version>0 AND authorization_version>0 AND authorization_clock>0
    ),
    CONSTRAINT analytics_dashboard_exports_state_check CHECK (
        state IN ('queued','running','ready','failed','cancelled','deleting','deleted','expired')
		AND row_revision>0 AND attempt_generation BETWEEN 0 AND 3
		AND delete_attempt_generation BETWEEN 0 AND 3
    ),
    CONSTRAINT analytics_dashboard_exports_time_check CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND pg_catalog.isfinite(expires_at) AND created_at<=updated_at
        AND expires_at>created_at
        AND (ready_at IS NULL OR pg_catalog.isfinite(ready_at))
        AND (deleted_at IS NULL OR pg_catalog.isfinite(deleted_at))
		AND (delete_verified_at IS NULL OR pg_catalog.isfinite(delete_verified_at))
    )
);

CREATE TABLE analytics_dashboard_export_attempt_receipts (
	export_id uuid NOT NULL,
	attempt_generation bigint NOT NULL,
	object_key text NOT NULL,
	key_id text NOT NULL,
	wrapped_dek bytea NOT NULL,
	wrap_nonce bytea NOT NULL,
	header_sha256 bytea NOT NULL,
	ciphertext_sha256 bytea NOT NULL,
	plaintext_sha256 bytea NOT NULL,
	ciphertext_bytes bigint NOT NULL,
	plaintext_bytes bigint NOT NULL,
	row_count bigint NOT NULL,
	outcome text NOT NULL,
	object_version text,
	provider_request_id text,
	started_at timestamptz NOT NULL,
	completed_at timestamptz,
	PRIMARY KEY(export_id,attempt_generation),
	FOREIGN KEY(export_id) REFERENCES analytics_dashboard_exports(export_id)
		ON UPDATE NO ACTION ON DELETE RESTRICT,
	CHECK(attempt_generation BETWEEN 1 AND 3),
	CHECK(outcome IN ('started','missing','ready','unknown','failed','deleted')),
	CHECK(pg_catalog.octet_length(wrapped_dek)=48
		AND pg_catalog.octet_length(wrap_nonce)=12
		AND pg_catalog.octet_length(header_sha256)=32
		AND pg_catalog.octet_length(ciphertext_sha256)=32
		AND pg_catalog.octet_length(plaintext_sha256)=32),
	CHECK(ciphertext_bytes>=35 AND plaintext_bytes>0 AND row_count>=0),
	CHECK(pg_catalog.isfinite(started_at)
		AND (completed_at IS NULL OR pg_catalog.isfinite(completed_at)))
);

CREATE INDEX analytics_dashboard_exports_due_idx
    ON analytics_dashboard_exports(state,updated_at,export_id)
    WHERE state IN ('queued','running','deleting','ready');
