SELECT *
FROM platform.billing_discount_offerings_snapshot_read_v1(
    $1::uuid, $2::uuid
);
