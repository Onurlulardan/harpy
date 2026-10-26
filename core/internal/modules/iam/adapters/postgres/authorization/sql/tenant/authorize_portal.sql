SELECT actor_document,resource_document,action_document,authorization_clock,
       bindings_document,implications_document
FROM iam_portal_authorization_evidence_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1::uuid,$2::uuid,$3::text,$4::text
);
