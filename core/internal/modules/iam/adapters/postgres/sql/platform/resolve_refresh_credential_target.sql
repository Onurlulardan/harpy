SELECT tenant_id, source_token_id, source_family_id, source_selector,
       source_revision, source_sha256,
       group_generation, iam_generation, layout_generation
FROM platform.iam_resolve_refresh_target_v2($1);
