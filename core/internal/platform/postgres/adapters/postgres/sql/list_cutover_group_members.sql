SELECT domain
FROM platform.cutover_runtime_lock_group_members_v1('tenant', $1, $2);
