//go:build integration

package postgres

func adoptedInvitationToken(
	tenant iamFoundationTenant, invitationID, tokenID string,
) map[string]any {
	return map[string]any{
		"id": tokenID, "purpose": "invitation_accept_v1", "user_id": nil,
		"identifier_id": nil, "identifier_kind": nil, "identifier_revision": nil,
		"invitation_id":         invitationID,
		"invitation_expires_at": "2026-07-23T01:00:00Z",
		"invitation_revision":   1, "portal_binding_id": nil,
		"portal_setup_revision": nil, "selector": adoptedSelector(tenant),
		"hash_key_id": "adoption-fixture-v1", "secret_hmac": adoptedHash("b3"),
		"binding_sha256": adoptedHash("b4"), "state": "pending",
		"attempt_count": 0, "revision": 1, "created_at": adoptedAt,
		"updated_at": adoptedAt, "expires_at": "2026-07-23T01:00:00Z",
		"consumed_at": nil, "revoked_at": nil, "expired_at": nil,
		"purge_after": "2026-07-24T01:00:00Z",
	}
}

func adoptedSelector(tenant iamFoundationTenant) string {
	if tenant.schema == "t_iam_b" {
		return `\xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb`
	}
	return `\xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`
}
