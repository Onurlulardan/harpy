//go:build integration

package postgres

func adoptedResourceTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedID(tenant, 8)
	return []iamV1AdoptionTarget{adoptedTarget(
		"resources", "resource", "iam_resources", id, map[string]any{
			"id": id, "resource_key": "iam.permission_requests", "owner_domain": "iam",
			"resource_kind": "request", "display_name": "Permission requests",
			"description": "V1 adoption rehearsal resource", "is_protected": true,
			"state": "active", "revision": 1, "created_at": adoptedAt,
			"updated_at": adoptedAt, "retired_at": nil,
		},
	)}
}

func adoptedActionTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	manageID, reviewID := adoptedID(tenant, 9), adoptedID(tenant, 19)
	return []iamV1AdoptionTarget{
		adoptedTarget("actions", "manage", "iam_actions", manageID, map[string]any{
			"id": manageID, "action_key": "manage", "display_name": "Manage",
			"description": "V1 adoption management authority", "is_protected": true,
			"state": "active", "revision": 1, "created_at": adoptedAt,
			"updated_at": adoptedAt, "retired_at": nil,
		}),
		adoptedTarget("actions", "review", "iam_actions", reviewID, map[string]any{
			"id": reviewID, "action_key": "review", "display_name": "Review",
			"description": "V1 adoption review authority", "is_protected": true,
			"state": "active", "revision": 1, "created_at": adoptedAt,
			"updated_at": adoptedAt, "retired_at": nil,
		}),
	}
}

func adoptedPermissionTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	permissionID, bindingID := adoptedID(tenant, 10), adoptedID(tenant, 11)
	reviewPermissionID := adoptedID(tenant, 20)
	return []iamV1AdoptionTarget{
		adoptedTarget("permissions", "manage-permission", "iam_permissions", permissionID,
			map[string]any{
				"id": permissionID, "match_kind": "exact",
				"resource_id": adoptedID(tenant, 8), "action_id": adoptedID(tenant, 9),
				"is_protected": true, "state": "active", "revision": 1,
				"created_at": adoptedAt, "updated_at": adoptedAt, "retired_at": nil,
			}),
		adoptedTarget("permissions", "user-binding", "iam_user_permission_bindings",
			bindingID, map[string]any{
				"id": bindingID, "user_id": adoptedID(tenant, 2),
				"permission_id": permissionID, "permission_match_kind": "exact",
				"state": "active", "revision": 1, "created_at": adoptedAt,
				"updated_at": adoptedAt, "activated_at": adoptedAt, "revoked_at": nil,
			}),
		adoptedTarget("permissions", "review-permission", "iam_permissions",
			reviewPermissionID, map[string]any{
				"id": reviewPermissionID, "match_kind": "exact",
				"resource_id": adoptedID(tenant, 8), "action_id": adoptedID(tenant, 19),
				"is_protected": true, "state": "active", "revision": 1,
				"created_at": adoptedAt, "updated_at": adoptedAt, "retired_at": nil,
			}),
	}
}
