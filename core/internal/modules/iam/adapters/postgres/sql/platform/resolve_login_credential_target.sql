SELECT tenant_id, source_user_id, source_identifier_id,
       source_identifier_kind, source_revision, source_sha256,
       group_generation, iam_generation, layout_generation
FROM platform.iam_resolve_login_target_v7($1, $2, $3, $4);
