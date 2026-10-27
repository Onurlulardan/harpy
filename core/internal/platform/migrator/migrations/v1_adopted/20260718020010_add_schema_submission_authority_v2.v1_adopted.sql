-- +goose Up
ALTER TABLE platform.schema_jobs
    DROP CONSTRAINT schema_jobs_plan_contract_version_check;

ALTER TABLE platform.schema_jobs
    ADD CONSTRAINT schema_jobs_plan_contract_version_check CHECK (
        plan_contract_version IN ('forms-schema-plan.v1','forms-schema-plan.v2')
    ),
    ADD COLUMN submission_authority_id uuid,
    ADD COLUMN submission_authority_group_id text,
    ADD COLUMN submission_group_generation bigint,
    ADD COLUMN submission_billing_generation bigint,
    ADD COLUMN submission_forms_generation bigint;

ALTER TABLE platform.schema_jobs
    ADD CONSTRAINT schema_jobs_submission_authority_v2_check CHECK (
        (plan_contract_version='forms-schema-plan.v1'
         AND submission_authority_id IS NULL
         AND submission_authority_group_id IS NULL
         AND submission_group_generation IS NULL
         AND submission_billing_generation IS NULL
         AND submission_forms_generation IS NULL)
        OR
        (plan_contract_version='forms-schema-plan.v2'
         AND compiler_version='2.0.0'
         AND canonical_plan_bytes IS NOT NULL
         AND pg_catalog.uuid_extract_version(submission_authority_id)
                IS NOT DISTINCT FROM 7
         AND submission_authority_group_id='billing_forms'
         AND submission_group_generation BETWEEN 1 AND 9007199254740991
         AND submission_billing_generation BETWEEN 1 AND 9007199254740991
         AND submission_forms_generation BETWEEN 1 AND 9007199254740991)
    );
