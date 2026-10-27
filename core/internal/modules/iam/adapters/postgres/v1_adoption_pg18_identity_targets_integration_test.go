//go:build integration

package postgres

import "strings"

func adoptedOrganizationTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedOrganizationID(tenant)
	marker := adoptedMarker(tenant)
	return []iamV1AdoptionTarget{adoptedTarget(
		"organizations", "organization", "iam_organizations", id, map[string]any{
			"id": id, "name": "Adopted Organization " + strings.ToUpper(marker),
			"slug": "adopted-" + marker, "status": "active", "parent_id": nil,
			"profile_resource_kind":     "organization",
			"profile_schema_version_id": iamOrganizationProfileSchemaID,
			"profile_values":            map[string]any{}, "profile_sha256": adoptedHash("ab"),
			"revision": 1, "created_at": adoptedAt, "updated_at": adoptedAt,
			"tombstoned_at": nil,
		},
	)}
}

func adoptedUserTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id := adoptedID(tenant, 2)
	return []iamV1AdoptionTarget{adoptedTarget(
		"users", "user", "iam_users", id, map[string]any{
			"id": id, "status": "active", "identity_kind": "human",
			"profile_resource_kind":     "user",
			"profile_schema_version_id": iamUserProfileSchemaID,
			"profile_values":            map[string]any{}, "profile_sha256": adoptedHash("ac"),
			"token_version": 1, "authorization_version": 0, "revision": 1,
			"created_at": adoptedAt, "updated_at": adoptedAt, "tombstoned_at": nil,
		},
	)}
}

func adoptedIdentifierTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	id, userID := adoptedID(tenant, 3), adoptedID(tenant, 2)
	email := "adopted-" + adoptedMarker(tenant) + "@example.test"
	return []iamV1AdoptionTarget{adoptedTarget(
		"identifiers", "email", "iam_user_identifiers", id, map[string]any{
			"id": id, "user_id": userID, "kind": "email_v1",
			"display_value": email, "normalized_value": email,
			"normalized_sha256": adoptedHash("ad"), "state": "active", "revision": 1,
			"verified_at": adoptedAt, "activated_at": adoptedAt,
			"tombstoned_at": nil, "cooldown_until": nil, "retired_at": nil,
			"created_at": adoptedAt, "updated_at": adoptedAt,
		},
	)}
}

func adoptedCredentialTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	userID := adoptedID(tenant, 2)
	return []iamV1AdoptionTarget{adoptedTarget(
		"credentials", "password", "iam_password_credentials", userID, map[string]any{
			"user_id": userID, "credential_version": 1, "state": "active",
			"algorithm": "bcrypt_v1_adopted", "salt": nil, "tag": nil,
			"memory_kib": nil, "iterations": nil, "parallelism": nil,
			"legacy_bcrypt_hash": "$2a$12$" + strings.Repeat("a", 53),
			"upgrade_required":   true, "revision": 1,
			"created_at": adoptedAt, "updated_at": adoptedAt,
			"last_used_at": nil, "invalidated_at": nil,
		},
	)}
}
