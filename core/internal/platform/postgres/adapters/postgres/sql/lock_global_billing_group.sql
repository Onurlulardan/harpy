SELECT state, writer, generation
FROM platform.cutover_runtime_lock_group_v1('global', $1, 'billing');
