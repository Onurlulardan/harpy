//go:build integration

package postgres

func adoptedMembershipTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedID(tenant, 4)
	return []iamV1AdoptionTarget{adoptedTarget(
		"memberships", "membership", "iam_organization_memberships", id,
		map[string]any{
			"id": id, "organization_id": adoptedID(tenant, 1),
			"user_id": adoptedID(tenant, 2), "membership_kind": "member",
			"state": "active", "revision": 1, "created_at": adoptedAt,
			"updated_at": adoptedAt, "tombstoned_at": nil,
		},
	)}
}

func adoptedRoleTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedID(tenant, 5)
	return []iamV1AdoptionTarget{adoptedTarget(
		"roles", "role", "iam_roles", id, map[string]any{
			"id": id, "scope_kind": "tenant", "scope_id": tenant.id,
			"organization_id": nil, "name": "Adopted Operator",
			"name_key":    `\x41646f70746564204f70657261746f72`,
			"description": "V1 adoption rehearsal role", "status": "active",
			"is_protected": false, "is_default": false,
			"profile_resource_kind":     "role",
			"profile_schema_version_id": iamRoleProfileSchemaID,
			"profile_values":            map[string]any{}, "profile_sha256": adoptedHash("ae"),
			"revision": 1, "created_at": adoptedAt, "updated_at": adoptedAt,
			"retired_at": nil,
		},
	)}
}

func adoptedAssignmentTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedID(tenant, 6)
	return []iamV1AdoptionTarget{adoptedTarget(
		"assignments", "assignment", "iam_user_role_assignments", id,
		map[string]any{
			"id": id, "user_id": adoptedID(tenant, 2),
			"role_id": adoptedID(tenant, 5), "role_scope_kind": "tenant",
			"role_scope_id": tenant.id, "organization_id": nil, "membership_id": nil,
			"state": "active", "revision": 1, "created_at": adoptedAt,
			"updated_at": adoptedAt, "revoked_at": nil,
		},
	)}
}

func adoptedPortalTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedID(tenant, 7)
	return []iamV1AdoptionTarget{adoptedTarget(
		"portal", "access-mode", "iam_user_access_modes", id, map[string]any{
			"id": id, "user_id": adoptedID(tenant, 2), "mode": "operator_console",
			"state": "active", "revision": 1, "created_at": adoptedAt,
			"updated_at": adoptedAt, "activated_at": adoptedAt,
			"suspended_at": nil, "retired_at": nil,
		},
	)}
}
