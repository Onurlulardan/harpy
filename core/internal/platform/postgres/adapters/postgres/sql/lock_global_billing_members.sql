SELECT domain
FROM platform.cutover_runtime_lock_group_members_v1('global', $1, 'billing');
