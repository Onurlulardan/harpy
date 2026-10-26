SELECT document, next_cursor
FROM platform.tenancy_customer_lifecycle_list_v1($1::uuid,$2::integer,$3::uuid)
