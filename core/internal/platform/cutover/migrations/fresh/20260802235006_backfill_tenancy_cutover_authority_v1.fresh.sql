-- +goose Up
-- Migration stream version: 20260802235006.
ALTER TABLE platform.domain_cutovers
    DROP CONSTRAINT domain_cutovers_domain_check,
    ADD CONSTRAINT domain_cutovers_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'forms', 'iam',
            'notifications', 'system_resource_layout', 'tenancy', 'work', 'workflow'
        )) NOT VALID;

ALTER TABLE platform.domain_cutovers
    VALIDATE CONSTRAINT domain_cutovers_domain_check;

ALTER TABLE platform.cutover_group_members
    DROP CONSTRAINT cutover_group_members_domain_check,
    ADD CONSTRAINT cutover_group_members_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'forms', 'iam',
            'notifications', 'system_resource_layout', 'tenancy', 'work', 'workflow'
        )) NOT VALID;

ALTER TABLE platform.cutover_group_members
    VALIDATE CONSTRAINT cutover_group_members_domain_check;

ALTER TABLE platform.cutover_transition_events
    DROP CONSTRAINT cutover_transition_events_domain_check,
    ADD CONSTRAINT cutover_transition_events_domain_check
        CHECK (
            domain IS NULL OR domain IN (
                'analytics', 'billing', 'catalog', 'documents', 'forms', 'iam',
                'notifications', 'system_resource_layout', 'tenancy', 'work', 'workflow'
            )
        ) NOT VALID;

ALTER TABLE platform.cutover_transition_events
    VALIDATE CONSTRAINT cutover_transition_events_domain_check;

ALTER TABLE platform.contract_retirements
    DROP CONSTRAINT contract_retirements_domain_check,
    ADD CONSTRAINT contract_retirements_domain_check
        CHECK (domain IN (
            'analytics', 'billing', 'catalog', 'documents', 'forms', 'iam',
            'notifications', 'system_resource_layout', 'tenancy', 'work', 'workflow'
        )) NOT VALID;

ALTER TABLE platform.contract_retirements
    VALIDATE CONSTRAINT contract_retirements_domain_check;

SELECT platform.cutover_ensure_tenancy_authority_v1(tenant.scope_id)
FROM (
    SELECT DISTINCT domain.scope_id
    FROM platform.domain_cutovers AS domain
    WHERE domain.scope_kind = 'tenant'
) AS tenant;

SELECT 1 / CASE WHEN EXISTS (
        SELECT 1
        FROM platform.domain_cutovers AS domain
        LEFT JOIN platform.cutover_groups AS authority
          ON authority.scope_kind = domain.scope_kind
         AND authority.scope_id = domain.scope_id
         AND authority.group_id = domain.domain
        LEFT JOIN platform.cutover_group_members AS member
          ON member.scope_kind = authority.scope_kind
         AND member.scope_id = authority.scope_id
         AND member.group_id = authority.group_id
         AND member.ordinal = 1
         AND member.domain = domain.domain
        WHERE domain.scope_kind = 'tenant'
          AND domain.domain = 'tenancy'
          AND (
              authority.scope_id IS NULL
              OR authority.state IS DISTINCT FROM domain.state
              OR authority.writer IS DISTINCT FROM domain.writer
              OR authority.generation IS DISTINCT FROM domain.generation
              OR member.scope_id IS NULL
          )
    ) THEN 0 ELSE 1 END;
