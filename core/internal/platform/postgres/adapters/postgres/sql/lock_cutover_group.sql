SELECT state, writer, generation
FROM platform.cutover_runtime_lock_group_v1('tenant', $1, $2);
