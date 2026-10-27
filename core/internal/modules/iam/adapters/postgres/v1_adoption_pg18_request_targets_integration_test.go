//go:build integration

package postgres

func adoptedPermissionRequestTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	groupID, requestID := adoptedID(tenant, 12), adoptedID(tenant, 13)
	permissionID := adoptedID(tenant, 10)
	return []iamV1AdoptionTarget{
		adoptedTarget("permission_requests", "group", "iam_permission_request_groups",
			groupID, map[string]any{
				"id": groupID, "requester_user_id": adoptedID(tenant, 2),
				"request_mode":                     "SELF_SERVICE",
				"creator_authority_permission_id":  permissionID,
				"reviewer_authority_permission_id": adoptedID(tenant, 20),
				"step_up_purpose_sha256":           adoptedHash("b1"), "state": "PENDING",
				"request_count": 1, "reviewer_user_id": nil, "review_task_id": nil,
				"revision": 1, "expires_at": "2026-07-17T01:00:00Z",
				"terminal_at": nil, "created_at": adoptedAt, "updated_at": adoptedAt,
			}),
		adoptedTarget("permission_requests", "request", "iam_permission_requests",
			requestID, map[string]any{
				"id": requestID, "group_id": groupID, "ordinal": 1,
				"target_kind": "USER", "target_user_id": adoptedID(tenant, 2),
				"target_role_id": nil, "target_organization_id": nil,
				"permission_id": permissionID, "permission_match_kind": "exact",
				"permission_is_protected": true,
				"planned_binding_id":      adoptedID(tenant, 18), "granted_binding_id": nil,
				"state": "PENDING", "revision": 1, "terminal_at": nil,
				"created_at": adoptedAt, "updated_at": adoptedAt,
			}),
	}
}

func adoptedInvitationTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	invitationID, tokenID := adoptedID(tenant, 14), adoptedID(tenant, 15)
	email := "invited-" + adoptedMarker(tenant) + "@example.test"
	return []iamV1AdoptionTarget{
		adoptedTarget("invitations", "invitation", "iam_invitations", invitationID,
			map[string]any{
				"id": invitationID, "identifier_kind": "email_v1",
				"display_value": email, "normalized_value": email,
				"normalized_sha256": adoptedHash("b2"),
				"role_id":           adoptedID(tenant, 5), "role_scope_kind": "tenant",
				"role_scope_id": tenant.id, "organization_id": nil,
				"invited_by": adoptedID(tenant, 2), "accepted_user_id": nil,
				"accepted_identifier_id": nil, "accepted_identifier_kind": nil,
				"state": "pending", "revision": 1, "created_at": adoptedAt,
				"updated_at": adoptedAt, "expires_at": "2026-07-23T01:00:00Z",
				"accepted_at": nil, "revoked_at": nil, "expired_at": nil,
			}),
		adoptedTarget("invitations", "token", "iam_one_time_tokens", tokenID,
			adoptedInvitationToken(tenant, invitationID, tokenID)),
	}
}
