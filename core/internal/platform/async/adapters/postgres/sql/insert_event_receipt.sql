INSERT INTO platform.inbox (
    id, record_kind, consumer_name, consumer_version,
    tenant_id, scope_kind, scope_id,
    producer_group_id, producer_group_generation,
    producer_domain, producer_generation, writer_origin,
    consumer_group_id, consumer_group_generation,
    consumer_domain, consumer_generation, consumer_writer_origin,
    aggregate_type, aggregate_id, event_id, event_type,
    event_version, event_sequence, receipt_sha256
) VALUES (
    $1::uuid, 'event_receipt', $2::text, $3::integer,
    $4::uuid, 'tenant', $4::uuid,
    $5::text, $6::bigint, $7::text, $8::bigint, $9::text,
    $10::text, $11::bigint, $12::text, $13::bigint, 'v2',
    $14::text, $15::uuid, $16::uuid, $17::text,
    $18::integer, $19::bigint, $20::text
);
