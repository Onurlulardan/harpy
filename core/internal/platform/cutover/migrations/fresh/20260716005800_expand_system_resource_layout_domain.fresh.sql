-- +goose Up
ALTER TABLE platform.domain_cutovers
    ADD CONSTRAINT domain_cutovers_authority_generation_unique
        UNIQUE (scope_kind, scope_id, domain, generation);

ALTER TABLE platform.cutover_groups
    ADD CONSTRAINT cutover_groups_authority_generation_unique
        UNIQUE (scope_kind, scope_id, group_id, generation);

ALTER TABLE platform.domain_cutovers
    DROP CONSTRAINT domain_cutovers_domain_check,
    ADD CONSTRAINT domain_cutovers_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'iam',
            'notifications', 'system_resource_layout', 'work', 'workflow'
        )) NOT VALID;

ALTER TABLE platform.domain_cutovers
    VALIDATE CONSTRAINT domain_cutovers_domain_check;

ALTER TABLE platform.cutover_group_members
    DROP CONSTRAINT cutover_group_members_domain_check,
    ADD CONSTRAINT cutover_group_members_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'iam',
            'notifications', 'system_resource_layout', 'work', 'workflow'
        )) NOT VALID;

ALTER TABLE platform.cutover_group_members
    VALIDATE CONSTRAINT cutover_group_members_domain_check;

ALTER TABLE platform.cutover_transition_events
    DROP CONSTRAINT cutover_transition_events_domain_check,
    ADD CONSTRAINT cutover_transition_events_domain_check
        CHECK (
            domain IS NULL OR domain IN (
                'analytics', 'billing', 'catalog', 'documents', 'iam',
                'notifications', 'system_resource_layout', 'work', 'workflow'
            )
        ) NOT VALID;

ALTER TABLE platform.cutover_transition_events
    VALIDATE CONSTRAINT cutover_transition_events_domain_check;

ALTER TABLE platform.contract_retirements
    DROP CONSTRAINT contract_retirements_domain_check,
    ADD CONSTRAINT contract_retirements_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'iam',
            'notifications', 'system_resource_layout', 'work', 'workflow'
        )) NOT VALID;

ALTER TABLE platform.contract_retirements
    VALIDATE CONSTRAINT contract_retirements_domain_check;
