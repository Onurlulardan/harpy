SELECT *
FROM platform.tenancy_customer_provision_claim_v1(
    $1::uuid,$2::text,$3::text,$4::integer
)
