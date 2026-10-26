SELECT aggregate_revision, authorization_clock, event_sequence
FROM iam_identity_change_audit_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::uuid, $5::bytea, $6::text, $7::text
);
