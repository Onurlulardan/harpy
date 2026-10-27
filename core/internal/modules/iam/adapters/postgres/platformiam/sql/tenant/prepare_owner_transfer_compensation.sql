SELECT operation_id, registry_binding_id, scope_id, generation,
       previous_head_sha256, state_sha256, proposal_sha256, status, replayed
FROM iam_owner_transfer_compensate_prepare_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::bytea,$3::uuid,$4::bigint,$5::bigint,$6::bigint,
    $7::bytea,$8::timestamptz,$9::text,$10::boolean,$11::boolean,$12::boolean,
    $13::uuid,$14::bytea,$15::bigint,$16::bytea,$17::text,$18::bytea,$19::timestamptz
);
