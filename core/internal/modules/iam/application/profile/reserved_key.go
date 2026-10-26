package profile

import "strings"

var commonReservedKeys = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "deleted_at": {},
	"revision": {}, "status": {}, "tenant_id": {}, "owner_id": {},
	"password": {}, "password_hash": {}, "token_version": {},
	"authorization_version": {}, "permissions": {}, "role_ids": {},
	"organization_ids": {}, "session_id": {}, "identity_kind": {},
}

var securityKeyTerms = map[string]struct{}{
	"auth": {}, "authorization": {}, "credential": {}, "grant": {},
	"mfa": {}, "password": {}, "permission": {}, "secret": {},
	"session": {}, "tenant": {}, "token": {},
}

func reservedFieldKey(kind ResourceKind, key string) bool {
	if _, reserved := commonReservedKeys[key]; reserved {
		return true
	}
	for _, term := range strings.Split(key, "_") {
		if _, reserved := securityKeyTerms[term]; reserved {
			return true
		}
	}
	switch kind {
	case ResourceUser:
		return key == "email" || key == "normalized_email" || key == "account_type" ||
			key == "access_mode" || key == "portal_binding_id"
	case ResourceOrganization:
		return key == "name" || key == "slug" || key == "parent_id" || key == "member_ids"
	case ResourceRole:
		return key == "name" || key == "scope_kind" || key == "scope_id" ||
			key == "is_protected" || key == "is_default"
	default:
		return true
	}
}
