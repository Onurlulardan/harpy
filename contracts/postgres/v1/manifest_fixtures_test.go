package postgresv1

func validMigrationManifest() map[string]any {
	return map[string]any{
		"contract_version": "postgres-migration-manifest.v1",
		"engine":           "goose.v3",
		"postgres_major":   18,
		"history_table":    "platform.platform_migration_history",
		"out_of_order":     false,
		"entries": []any{map[string]any{
			"owner": "tenancy", "scope": "platform", "version": 20260714010000,
			"transaction_mode": "transactional", "phase": "foundation",
			"target_database_generation": 1, "next_minimum_runtime_generation": 1,
			"variants": map[string]any{
				"fresh": map[string]any{
					"path":   "core/internal/modules/tenancy/migrations/20260714010000_foundation.fresh.sql",
					"sha256": digest("a"),
				},
				"v1_adopted": map[string]any{
					"path":   "core/internal/modules/tenancy/migrations/20260714010000_foundation.v1_adopted.sql",
					"sha256": digest("b"),
				},
			},
		}},
	}
}

func validExtensionManifest() map[string]any {
	return map[string]any{
		"contract_version":        "postgres-extension-manifest.v1",
		"database_origin":         "fresh",
		"postgres_image_digest":   postgresImageDigest,
		"postgres_engine_version": "18.4",
		"control_path_policy": map[string]any{
			"resolver": "pg_config_sharedir", "relative_directory": "extension",
			"allow_absolute_paths": false, "allow_symlinks": false,
		},
		"extensions": []any{
			map[string]any{
				"name": "pg_trgm", "version": "1.6", "install": true,
				"schema": "platform_ext", "owner": "harpy_extension_admin",
				"control_file": "pg_trgm.control", "control_file_sha256": digest("c"),
				"capabilities": []any{"trigram_search"},
			},
			map[string]any{
				"name": "vector", "version": "0.8.3", "install": true,
				"schema": "platform_ext", "owner": "harpy_extension_admin",
				"control_file": "vector.control", "control_file_sha256": digest("d"),
				"capabilities": []any{"vector_similarity"},
			},
		},
	}
}

func validRolesManifest(origin string) map[string]any {
	names := []string{
		"harpy_owner", "harpy_cutover_authority_owner",
		"harpy_runtime", "harpy_v2_runtime",
		"harpy_migrator", "harpy_cluster_admin", "harpy_extension_admin",
		"harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer",
		"harpy_backup_reader", "harpy_backup", "harpy_restore",
		"harpy_reactivation_observer",
		"harpy_tenant_migration_executor",
	}
	if origin == "v1_adopted" {
		names = append(names, "harpy_v1_runtime", "harpy_legacy_extension_owner")
	}
	roles := make([]any, 0, len(names))
	for _, name := range names {
		roles = append(roles, validRole(name))
	}
	return map[string]any{
		"contract_version": "postgres-roles-manifest.v1",
		"database_origin":  origin, "roles": roles,
	}
}

func validRole(name string) map[string]any {
	login := name == "harpy_v2_runtime" || name == "harpy_migrator" ||
		name == "harpy_v1_runtime" || name == "harpy_backup" ||
		name == "harpy_restore" || name == "harpy_reactivation_observer"
	return map[string]any{
		"name": name, "login": login, "inherit": false,
		"superuser":   name == "harpy_extension_admin",
		"createdb":    name == "harpy_cluster_admin",
		"createrole":  name == "harpy_cluster_admin",
		"replication": false, "bypassrls": false,
		"connection_limit": roleConnectionLimit(name),
		"password_state":   "null", "valid_until": "infinity",
	}
}

func roleConnectionLimit(name string) int {
	switch name {
	case "harpy_cluster_admin", "harpy_extension_admin":
		return 1
	case "harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer":
		return 0
	default:
		return -1
	}
}
