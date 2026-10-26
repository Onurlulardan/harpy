SELECT *
FROM platform.billing_discount_eligibility_snapshot_read_v1(
    $1::uuid, $2::uuid, $3::text, $4::uuid, $5::timestamptz, $6::uuid
)
AS snapshot(
    tenant_id uuid,
    candidate_count bigint,
    discount_id uuid,
    revision bigint,
    discount_version_id uuid,
    version bigint,
    subscription_id uuid,
    offering_version_id uuid,
    price_version_id uuid,
    subscription_currency text,
    normalized_code text,
    discount_kind text,
    fixed_minor_units bigint,
    currency text,
    currency_registry_version bigint,
    basis_points integer,
    duration text,
    repeating_cycles bigint,
    priority integer,
    global_use_limit bigint,
    per_tenant_use_limit bigint,
    all_offerings boolean,
    expected_offering_count smallint,
    valid_start timestamptz,
    valid_end timestamptz,
    available boolean,
    global_used bigint,
    tenant_used bigint,
    offering_allowed boolean,
    already_redeemed boolean
);
