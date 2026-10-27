-- +goose Up

ALTER TABLE iam_profile_backfill_plans
    ADD COLUMN recovery_of_plan_id uuid,
    ADD COLUMN recovery_root_plan_id uuid,
    ADD COLUMN recovery_sequence bigint,
    ADD COLUMN parent_plan_sha256 bytea,
    ADD COLUMN target_definition_sha256 bytea,
    ADD COLUMN inherited_target_count bigint,
    ADD COLUMN inherited_target_sha256 bytea,
    ADD COLUMN remaining_source_count bigint,
    ADD COLUMN remaining_source_sha256 bytea,
    ADD COLUMN expected_remaining_target_sha256 bytea,
    ADD COLUMN initial_rolling_sha256 bytea,
    ADD CONSTRAINT iam_profile_backfill_plans_recovery_parent_fk
        FOREIGN KEY (recovery_of_plan_id)
        REFERENCES iam_profile_backfill_plans (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT iam_profile_backfill_plans_recovery_root_fk
        FOREIGN KEY (recovery_root_plan_id)
        REFERENCES iam_profile_backfill_plans (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT iam_profile_backfill_plans_recovery_presence_check CHECK (
        (recovery_of_plan_id IS NULL AND recovery_root_plan_id IS NULL AND
         recovery_sequence IS NULL AND parent_plan_sha256 IS NULL AND
         target_definition_sha256 IS NULL AND inherited_target_count IS NULL AND
         inherited_target_sha256 IS NULL AND remaining_source_count IS NULL AND
         remaining_source_sha256 IS NULL AND
         expected_remaining_target_sha256 IS NULL AND
         initial_rolling_sha256 IS NULL) OR
        (recovery_of_plan_id IS NOT NULL AND recovery_root_plan_id IS NOT NULL AND
         recovery_sequence IS NOT NULL AND parent_plan_sha256 IS NOT NULL AND
         target_definition_sha256 IS NOT NULL AND
         inherited_target_count IS NOT NULL AND
         inherited_target_sha256 IS NOT NULL AND
         remaining_source_count IS NOT NULL AND
         remaining_source_sha256 IS NOT NULL AND
         expected_remaining_target_sha256 IS NOT NULL AND
         initial_rolling_sha256 IS NOT NULL)
    ),
    ADD CONSTRAINT iam_profile_backfill_plans_recovery_identity_check CHECK (
        recovery_of_plan_id IS NULL OR
        (recovery_of_plan_id <> id AND recovery_root_plan_id <> id AND
         recovery_sequence > 0)
    ),
    ADD CONSTRAINT iam_profile_backfill_plans_recovery_digest_check CHECK (
        recovery_of_plan_id IS NULL OR
        (pg_catalog.octet_length(parent_plan_sha256) = 32 AND
         pg_catalog.octet_length(target_definition_sha256) = 32 AND
         pg_catalog.octet_length(inherited_target_sha256) = 32 AND
         pg_catalog.octet_length(remaining_source_sha256) = 32 AND
         pg_catalog.octet_length(expected_remaining_target_sha256) = 32 AND
         pg_catalog.octet_length(initial_rolling_sha256) = 32)
    ),
    ADD CONSTRAINT iam_profile_backfill_plans_recovery_count_check CHECK (
        recovery_of_plan_id IS NULL OR
        (inherited_target_count >= 0 AND remaining_source_count >= 0 AND
         source_count = remaining_source_count AND
         expected_source_sha256 = remaining_source_sha256 AND
         expected_target_sha256 = expected_remaining_target_sha256 AND
         ((remaining_source_count = 0 AND page_count = 0 AND
           remaining_source_sha256 = pg_catalog.sha256(''::bytea) AND
           expected_remaining_target_sha256 = initial_rolling_sha256) OR
          (remaining_source_count > 0 AND page_count BETWEEN 1 AND 65536)))
    );

ALTER TABLE iam_profile_backfill_plans
    DROP CONSTRAINT iam_profile_backfill_plans_page_check,
    ADD CONSTRAINT iam_profile_backfill_plans_page_check CHECK (
        source_count >= 0 AND page_size BETWEEN 1 AND 250 AND revision > 0 AND
        ((recovery_of_plan_id IS NULL AND page_count >= 1) OR
         (recovery_of_plan_id IS NOT NULL AND page_count >= 0))
    );

CREATE UNIQUE INDEX iam_profile_backfill_plans_recovery_parent_uidx
    ON iam_profile_backfill_plans (recovery_of_plan_id)
    WHERE recovery_of_plan_id IS NOT NULL;

CREATE UNIQUE INDEX iam_profile_backfill_plans_recovery_sequence_uidx
    ON iam_profile_backfill_plans (recovery_root_plan_id, recovery_sequence)
    WHERE recovery_root_plan_id IS NOT NULL;

CREATE INDEX iam_profile_backfill_plans_recovery_state_idx
    ON iam_profile_backfill_plans (
        resource_kind, target_schema_id, state, recovery_sequence DESC
    );

CREATE VIEW iam_profile_recovery_resources_v1 AS
SELECT 'user'::text AS resource_kind, usr.id AS resource_id,
       usr.profile_schema_version_id AS schema_version_id,
       usr.revision AS resource_revision, usr.profile_values,
       usr.profile_sha256
FROM iam_users AS usr
UNION ALL
SELECT 'organization', organization.id,
       organization.profile_schema_version_id, organization.revision,
       organization.profile_values, organization.profile_sha256
FROM iam_organizations AS organization
UNION ALL
SELECT 'role', role.id, role.profile_schema_version_id, role.revision,
       role.profile_values, role.profile_sha256
FROM iam_roles AS role;
