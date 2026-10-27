package postgres

import "github.com/google/uuid"

type OutboxAuthority struct {
	owner           *workerIdentity
	scopeKind       string
	scopeID         uuid.UUID
	groupID         string
	groupGeneration int64
	domain          string
	generation      int64
}

func newOutboxAuthority(
	owner *workerIdentity,
	scopeKind string,
	scopeID uuid.UUID,
	groupID string,
	groupGeneration int64,
	domain string,
	generation int64,
) (OutboxAuthority, error) {
	authority := OutboxAuthority{
		owner: owner, scopeKind: scopeKind, scopeID: scopeID,
		groupID: groupID, groupGeneration: groupGeneration,
		domain: domain, generation: generation,
	}
	if !authority.valid(owner) {
		return OutboxAuthority{}, ErrOutboxAuthorityInvalid
	}
	return authority, nil
}

func (authority OutboxAuthority) valid(owner *workerIdentity) bool {
	return owner != nil && owner.marker == 1 && authority.owner == owner &&
		(authority.scopeKind == "tenant" || authority.scopeKind == "global") &&
		authority.scopeID != uuid.Nil && authority.groupID != "" &&
		authority.groupGeneration > 0 && authority.domain != "" &&
		authority.generation > 0
}

func outboxAuthorityArguments(authority OutboxAuthority) []any {
	return []any{
		authority.scopeKind, authority.scopeID, authority.groupID,
		authority.groupGeneration, authority.domain, authority.generation,
	}
}
