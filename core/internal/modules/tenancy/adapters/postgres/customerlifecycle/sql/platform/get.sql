SELECT id,name,slug,status,lifecycle_revision,settings_revision,
       created_at,updated_at,offboarded_at,purged_at
FROM platform.tenancy_customer_lifecycle_get_v1($1::uuid,$2::uuid)
