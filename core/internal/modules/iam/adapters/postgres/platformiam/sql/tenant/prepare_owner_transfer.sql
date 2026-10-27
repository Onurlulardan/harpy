SELECT operation_id, registry_binding_id, scope_id, generation,
       previous_head_sha256, state_sha256, proposal_sha256, status, replayed
FROM iam_owner_transfer_prepare_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::bytea,$3::uuid,$4::bigint,$5::bigint,$6::bigint,$7::bigint,
    $8::bytea,$9::timestamptz,$10::text,$11::boolean,$12::boolean,$13::boolean,
    $14::uuid,$15::bigint,$16::bigint,$17::bigint,$18::uuid,$19::uuid,$20::uuid,
    $21::bigint,$22::bigint,$23::bytea,$24::bytea,$25::bytea,$26::text,
    $27::bytea,$28::timestamptz
);
