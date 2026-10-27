-- +goose Up
-- +goose StatementBegin
CREATE TABLE platform.support_contact_messages (
    id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(id) = 7),
    name text NOT NULL CHECK (pg_catalog.octet_length(name) BETWEEN 1 AND 200),
    email text NOT NULL CHECK (pg_catalog.octet_length(email) BETWEEN 3 AND 255),
    company text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(company) <= 200),
    phone text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(phone) <= 50),
    employee_count text NOT NULL DEFAULT '' CHECK (employee_count IN ('','1-10','11-50','51-200','201-500','500+')),
    message text NOT NULL CHECK (pg_catalog.octet_length(message) BETWEEN 1 AND 2000),
    contact_type text NOT NULL CHECK (contact_type IN ('general','demo_request','pricing_inquiry','technical_question','partnership','enterprise_plan','on_premise')),
    status text NOT NULL CHECK (status IN ('new','read','replied','archived')),
    source text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(source) <= 50),
    notes text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(notes) <= 8000),
    replied_at timestamp(6) with time zone,
    replied_by uuid,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamp(6) with time zone NOT NULL,
    updated_at timestamp(6) with time zone NOT NULL,
    CHECK ((replied_at IS NULL) = (replied_by IS NULL))
);

CREATE INDEX support_contact_messages_list_idx
ON platform.support_contact_messages (created_at DESC, id DESC);
CREATE INDEX support_contact_messages_status_idx
ON platform.support_contact_messages (status, created_at DESC, id DESC);

CREATE TABLE platform.support_bug_reports (
    id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(id) = 7),
    tenant_id uuid NOT NULL,
    tenant_slug text NOT NULL CHECK (pg_catalog.octet_length(tenant_slug) BETWEEN 1 AND 100),
    tenant_name text NOT NULL CHECK (pg_catalog.octet_length(tenant_name) BETWEEN 1 AND 200),
    reporter_id uuid NOT NULL,
    reporter_name text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(reporter_name) <= 200),
    reporter_email text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(reporter_email) <= 255),
    title text NOT NULL CHECK (pg_catalog.octet_length(title) BETWEEN 3 AND 500),
    description text NOT NULL CHECK (pg_catalog.octet_length(description) BETWEEN 10 AND 8000),
    category text NOT NULL CHECK (category IN ('bug','feature_request','performance','ui_issue','data_issue','other')),
    severity text NOT NULL CHECK (severity IN ('critical','high','medium','low')),
    priority text NOT NULL CHECK (priority IN ('urgent','high','normal','low')),
    page_url text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(page_url) <= 2048),
    user_agent text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(user_agent) <= 1024),
    browser_info jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (pg_catalog.jsonb_typeof(browser_info) = 'object' AND pg_catalog.octet_length(browser_info::text) <= 16384),
    console_errors jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (pg_catalog.jsonb_typeof(console_errors) = 'array' AND pg_catalog.jsonb_array_length(console_errors) <= 50 AND pg_catalog.octet_length(console_errors::text) <= 32768),
    network_info jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (pg_catalog.jsonb_typeof(network_info) = 'array' AND pg_catalog.jsonb_array_length(network_info) <= 50 AND pg_catalog.octet_length(network_info::text) <= 32768),
    steps_to_reproduce text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(steps_to_reproduce) <= 8000),
    status text NOT NULL CHECK (status IN ('open','in_progress','resolved','closed','wont_fix','duplicate')),
    resolution text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(resolution) <= 8000),
    admin_notes text NOT NULL DEFAULT '' CHECK (pg_catalog.octet_length(admin_notes) <= 8000),
    resolved_at timestamp(6) with time zone,
    closed_at timestamp(6) with time zone,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamp(6) with time zone NOT NULL,
    updated_at timestamp(6) with time zone NOT NULL
);

CREATE INDEX support_bug_reports_tenant_list_idx
ON platform.support_bug_reports (tenant_id, created_at DESC, id DESC);
CREATE INDEX support_bug_reports_reporter_list_idx
ON platform.support_bug_reports (tenant_id, reporter_id, created_at DESC, id DESC);
CREATE INDEX support_bug_reports_global_list_idx
ON platform.support_bug_reports (created_at DESC, id DESC);

CREATE TABLE platform.support_command_receipts (
    receipt_id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(receipt_id) = 7),
    operation text NOT NULL CHECK (pg_catalog.octet_length(operation) BETWEEN 1 AND 100),
    principal_partition text NOT NULL CHECK (pg_catalog.octet_length(principal_partition) BETWEEN 1 AND 128),
    idempotency_key text NOT NULL CHECK (pg_catalog.octet_length(idempotency_key) BETWEEN 16 AND 200),
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    result_id uuid NOT NULL,
    result_revision bigint NOT NULL CHECK (result_revision > 0),
    response_document jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(response_document) = 'object' AND pg_catalog.octet_length(response_document::text) <= 262144),
    created_at timestamp(6) with time zone NOT NULL,
    UNIQUE (operation, principal_partition, idempotency_key)
);

CREATE TABLE platform.support_public_rate_leases (
    rate_key_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(rate_key_sha256) = 32),
    bucket_start timestamp(6) with time zone NOT NULL,
    request_count integer NOT NULL CHECK (request_count BETWEEN 1 AND 100000),
    updated_at timestamp(6) with time zone NOT NULL,
    PRIMARY KEY (rate_key_sha256, bucket_start)
);

CREATE TABLE platform.support_adoption_runs (
    run_id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(run_id) = 7),
    source_kind text NOT NULL CHECK (source_kind IN ('contact_global','bug_tenant')),
    source_tenant_id uuid,
    source_schema text NOT NULL CHECK (pg_catalog.octet_length(source_schema) BETWEEN 1 AND 128),
    cursor_id uuid,
    copied_count bigint NOT NULL DEFAULT 0 CHECK (copied_count >= 0),
    source_sha256 bytea,
    target_sha256 bytea,
    status text NOT NULL CHECK (status IN ('planned','applying','reconciling','completed','fenced')),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamp(6) with time zone NOT NULL,
    updated_at timestamp(6) with time zone NOT NULL,
    CHECK (source_sha256 IS NULL OR pg_catalog.octet_length(source_sha256) = 32),
    CHECK (target_sha256 IS NULL OR pg_catalog.octet_length(target_sha256) = 32)
);

REVOKE ALL ON TABLE platform.support_contact_messages FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON TABLE platform.support_bug_reports FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON TABLE platform.support_command_receipts FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON TABLE platform.support_public_rate_leases FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON TABLE platform.support_adoption_runs FROM PUBLIC, forgecrud_runtime;
-- +goose StatementEnd
