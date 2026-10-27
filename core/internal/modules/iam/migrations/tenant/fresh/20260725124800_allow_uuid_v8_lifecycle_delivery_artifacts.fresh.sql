-- +goose Up
ALTER TABLE iam_lifecycle_delivery_artifacts
    DROP CONSTRAINT iam_lifecycle_delivery_artifacts_ref_check;

ALTER TABLE iam_lifecycle_delivery_artifacts
    ADD CONSTRAINT iam_lifecycle_delivery_artifacts_ref_check CHECK (
        artifact_ref ~ '^iam-lifecycle-delivery://tenant/[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}/artifact/[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
    );
