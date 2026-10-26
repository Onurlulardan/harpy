SELECT
    procedure_contract_id,
    procedure_oid,
    procedure_signature,
    already_applied,
    applied_count,
    key_state_sha256,
    state_sha256
FROM platform.cutover_prepare_batch_v1(
    $1::uuid, $2::uuid, $3::text, $4::uuid, $5::smallint,
    $6::text, $7::integer, $8::pg_lsn, $9::xid8, $10::text, $11::text
)
