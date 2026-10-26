SELECT tenant_id, group_generation, $2::bytea,
       group_generation, iam_generation, layout_generation
FROM platform.iam_resolve_tenant_target_v1($1);
