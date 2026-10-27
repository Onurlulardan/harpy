SELECT * FROM iam_portal_subject_links_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::bigint,$4::bigint,$5::bigint,
    $6::uuid,$7::uuid,$8::uuid[]
);
