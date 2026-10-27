-- +goose Up
CREATE TABLE iam_v1_adoption_runs (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    contract_version text NOT NULL,
    source_sha256 bytea NOT NULL,
    mapping_sha256 bytea NOT NULL,
    plan_sha256 bytea NOT NULL,
    state text NOT NULL,
    eligible boolean NOT NULL,
    issue_count integer NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    projected_at timestamptz,
    verified_at timestamptz,
    CONSTRAINT iam_v1_adoption_runs_pkey PRIMARY KEY (operation_id, run_id),
    CONSTRAINT iam_v1_adoption_runs_contract_check CHECK (
        contract_version = 'harpy/iam-v1-adoption-plan/v1'
    ),
    CONSTRAINT iam_v1_adoption_runs_digest_check CHECK (
        pg_catalog.octet_length(source_sha256) = 32 AND
        pg_catalog.octet_length(mapping_sha256) = 32 AND
        pg_catalog.octet_length(plan_sha256) = 32
    ),
    CONSTRAINT iam_v1_adoption_runs_state_check CHECK (
        state IN ('blocked', 'planned', 'copying', 'projected', 'verified')
    ),
    CONSTRAINT iam_v1_adoption_runs_issue_check CHECK (
        issue_count >= 0 AND eligible = (issue_count = 0) AND
        eligible = (state <> 'blocked')
    ),
    CONSTRAINT iam_v1_adoption_runs_time_check CHECK (
        updated_at >= created_at AND
        (projected_at IS NULL OR projected_at BETWEEN created_at AND updated_at) AND
        (verified_at IS NULL OR verified_at BETWEEN created_at AND updated_at)
    )
);

CREATE TABLE iam_v1_adoption_category_evidence (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    category text NOT NULL,
    source_count bigint NOT NULL,
    source_sha256 bytea NOT NULL,
    staged_count bigint,
    staged_sha256 bytea,
    target_count bigint,
    target_sha256 bytea,
    relation_ordinal smallint,
    copy_state text NOT NULL DEFAULT 'planned',
    last_batch_ordinal integer,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_v1_adoption_category_evidence_pkey
        PRIMARY KEY (operation_id, run_id, category),
    CONSTRAINT iam_v1_adoption_category_evidence_run_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES iam_v1_adoption_runs (operation_id, run_id),
    CONSTRAINT iam_v1_adoption_category_evidence_category_check CHECK (
        category IN (
            'organizations', 'roles', 'users', 'memberships', 'assignments',
            'identifiers', 'portal', 'credentials', 'resources', 'actions',
            'permissions', 'permission_requests', 'invitations', 'profiles', 'layouts'
        )
    ),
    CONSTRAINT iam_v1_adoption_category_evidence_count_check CHECK (
        source_count >= 0 AND (staged_count IS NULL OR staged_count >= 0) AND
        (target_count IS NULL OR target_count >= 0)
    ),
    CONSTRAINT iam_v1_adoption_category_evidence_digest_check CHECK (
        pg_catalog.octet_length(source_sha256) = 32 AND
        (staged_sha256 IS NULL OR pg_catalog.octet_length(staged_sha256) = 32) AND
        (target_sha256 IS NULL OR pg_catalog.octet_length(target_sha256) = 32)
    ),
    CONSTRAINT iam_v1_adoption_category_evidence_copy_check CHECK (
        copy_state IN ('planned', 'initial_copy', 'delta', 'complete')
    )
);

CREATE TABLE iam_v1_adoption_issues (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    ordinal integer NOT NULL,
    code text NOT NULL,
    category text NOT NULL,
    source_key text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_v1_adoption_issues_pkey PRIMARY KEY (operation_id, run_id, ordinal),
    CONSTRAINT iam_v1_adoption_issues_run_fk FOREIGN KEY (operation_id, run_id)
        REFERENCES iam_v1_adoption_runs (operation_id, run_id),
    CONSTRAINT iam_v1_adoption_issues_ordinal_check CHECK (ordinal > 0),
    CONSTRAINT iam_v1_adoption_issues_value_check CHECK (
        code <> '' AND category <> '' AND pg_catalog.char_length(source_key) <= 512
    )
);
