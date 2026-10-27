-- +goose Up
CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    ordinal smallint NOT NULL,
    domain text NOT NULL,
    PRIMARY KEY (scope_kind, scope_id, group_id, ordinal),
    CONSTRAINT cutover_group_members_group_fk
        FOREIGN KEY (scope_kind, scope_id, group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_group_members_domain_fk
        FOREIGN KEY (scope_kind, scope_id, domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_group_members_ordinal_check CHECK (ordinal BETWEEN 1 AND 2),
    CONSTRAINT cutover_group_members_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'forms', 'iam',
            'notifications', 'work', 'workflow'
        )),
    CONSTRAINT cutover_group_members_group_domain_unique
        UNIQUE (scope_kind, scope_id, group_id, domain),
    CONSTRAINT cutover_group_members_domain_unique
        UNIQUE (scope_kind, scope_id, domain)
);

CREATE INDEX cutover_group_members_domain_idx
    ON platform.cutover_group_members (scope_kind, scope_id, domain, group_id);
