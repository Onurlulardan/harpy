-- +goose Up
ALTER TABLE work_v1_adoption_stage
    ADD CONSTRAINT work_v1_adoption_stage_target_digest_check
    CHECK (
        target_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(target_record::text, 'UTF8')
        )
    );
