-- +goose Up
UPDATE iam_profile_field_revisions AS snapshot
SET field_key = field.field_key,
    value_kind = field.value_kind,
    cardinality = field.cardinality,
    required = field.required,
    classification = field.classification,
    normalization_profile = field.normalization_profile,
    field_state = field.state,
    semantic_sha256 = iam_profile_semantic_sha256_v1(
        field.id, field.resource_kind, field.field_key, field.value_kind,
        field.cardinality, field.required, field.classification,
        field.normalization_profile
    )
FROM iam_profile_fields AS field
WHERE field.id = snapshot.field_id
  AND field.resource_kind = snapshot.resource_kind;
