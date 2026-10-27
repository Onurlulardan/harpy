package postgresv1

func validPrivilegesManifest(origin string) map[string]any {
	memberships := []any{
		map[string]any{
			"member": "harpy_v2_runtime", "role": "harpy_runtime",
			"admin_option": false, "inherit_option": false, "set_option": true,
		},
		map[string]any{
			"member": "harpy_migrator", "role": "harpy_owner",
			"admin_option": false, "inherit_option": false, "set_option": true,
		},
		map[string]any{
			"member": "harpy_backup", "role": "harpy_backup_reader",
			"admin_option": false, "inherit_option": false, "set_option": true,
		},
		map[string]any{
			"member": "harpy_restore", "role": "harpy_owner",
			"admin_option": false, "inherit_option": false, "set_option": true,
		},
	}
	executor := "cluster_bootstrap"
	if origin == "v1_adopted" {
		executor = "external_root_provisioner"
		memberships = append(memberships, map[string]any{
			"member": "harpy_v1_runtime", "role": "harpy_runtime",
			"admin_option": false, "inherit_option": true, "set_option": false,
		})
	}
	return map[string]any{
		"contract_version": "postgres-privileges-manifest.v1",
		"database_origin":  origin, "database_name": "harpy",
		"public_database_privileges": []any{}, "public_schema_create": false,
		"memberships": memberships,
		"grants": []any{
			databaseConnectGrant("harpy_backup"),
			databaseConnectGrant("harpy_cluster_admin"),
			databaseConnectGrant("harpy_extension_admin"),
			databaseConnectGrant("harpy_migrator"),
			databaseConnectGrant("harpy_reactivation_observer"),
			databaseConnectGrant("harpy_restore"),
			databaseConnectGrant("harpy_runtime"),
			databaseConnectGrant("harpy_v2_runtime"),
		},
		"default_privileges": []any{map[string]any{
			"owner": "harpy_owner", "schema_name": "platform",
			"object_kind": "table", "grantee": "harpy_runtime",
			"privileges":   []any{"SELECT", "INSERT", "UPDATE", "DELETE"},
			"grant_option": false,
		}},
		"executor_decisions": map[string]any{
			"bootstrap_role":        "harpy_cluster_admin",
			"extension_role":        "harpy_extension_admin",
			"runtime_session_role":  "harpy_v2_runtime",
			"runtime_current_role":  "harpy_runtime",
			"migrator_session_role": "harpy_migrator",
			"migrator_current_role": "harpy_owner",
			"ownership_executor":    executor,
		},
	}
}

func databaseConnectGrant(grantee string) map[string]any {
	return map[string]any{
		"object": map[string]any{
			"kind": "database", "database_name": "harpy", "object_name": "harpy",
		},
		"grantor": "harpy_owner", "grantee": grantee,
		"privileges": []any{"CONNECT"}, "grant_option": false,
	}
}

func validOwnershipManifest(origin, format string) map[string]any {
	executor := "cluster_bootstrap"
	if origin == "v1_adopted" {
		executor = "external_root_provisioner"
	}
	manifest := map[string]any{
		"contract_version": "postgres-ownership-manifest.v1",
		"manifest_id":      "11111111-1111-4111-8111-111111111111",
		"database_origin":  origin, "database_name": "harpy",
		"ownership_executor": executor, "format": format, "object_count": 1,
	}
	if format == "inline" {
		manifest["objects"] = []any{validOwnershipObject(executor)}
	} else {
		manifest["chunks"] = []any{map[string]any{
			"ordinal": 0, "count": 1, "length_bytes": 1024,
			"sha256": digest("e"), "path": "ownership/" + digest("e") + ".json",
		}}
	}
	return manifest
}

func validOwnershipChunk(origin string) map[string]any {
	executor := "cluster_bootstrap"
	if origin == "v1_adopted" {
		executor = "external_root_provisioner"
	}
	return map[string]any{
		"contract_version": "postgres-ownership-chunk.v1",
		"manifest_id":      "11111111-1111-4111-8111-111111111111",
		"database_origin":  origin, "database_name": "harpy",
		"ordinal": 0, "count": 1, "object_count": 1,
		"objects": []any{validOwnershipObject(executor)},
	}
}

func validOwnershipObject(executor string) map[string]any {
	return map[string]any{
		"object": map[string]any{
			"kind": "database", "database_name": "harpy", "object_name": "harpy",
		},
		"owner": "harpy_owner", "classification": "first_party", "executor": executor,
	}
}
