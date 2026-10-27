package postgres

import "github.com/google/uuid"

type outboxAuthorityCursor struct {
	present   bool
	scopeKind string
	scopeID   uuid.UUID
	groupID   string
	domain    string
}

type OutboxAuthorityPageToken struct {
	owner  *workerIdentity
	limit  int
	cursor outboxAuthorityCursor
}

type OutboxAuthorityPage struct {
	authorities []OutboxAuthority
	next        OutboxAuthorityPageToken
	hasNext     bool
}

func (repository *workerRepository) BeginOutboxAuthorityPage(
	limit int,
) (OutboxAuthorityPageToken, error) {
	if !repository.valid() {
		return OutboxAuthorityPageToken{}, ErrWorkerStoreUnavailable
	}
	if limit < 1 || limit > maximumOutboxAuthorityPageSize {
		return OutboxAuthorityPageToken{}, ErrOutboxAuthorityPageLimit
	}
	return OutboxAuthorityPageToken{owner: repository.identity, limit: limit}, nil
}

func (token OutboxAuthorityPageToken) valid(repository *workerRepository) bool {
	if !repository.valid() || token.owner != repository.identity ||
		token.limit < 1 || token.limit > maximumOutboxAuthorityPageSize {
		return false
	}
	if !token.cursor.present {
		return token.cursor == (outboxAuthorityCursor{})
	}
	return (token.cursor.scopeKind == "tenant" || token.cursor.scopeKind == "global") &&
		token.cursor.scopeID != uuid.Nil && token.cursor.groupID != "" &&
		token.cursor.domain != ""
}

func (page OutboxAuthorityPage) Authorities() []OutboxAuthority {
	return append([]OutboxAuthority(nil), page.authorities...)
}

func (page OutboxAuthorityPage) Next() (OutboxAuthorityPageToken, bool) {
	return page.next, page.hasNext
}

func makeOutboxAuthorityPage(
	token OutboxAuthorityPageToken,
	authorities []OutboxAuthority,
) OutboxAuthorityPage {
	page := OutboxAuthorityPage{
		authorities: append([]OutboxAuthority(nil), authorities...),
	}
	if len(page.authorities) <= token.limit {
		return page
	}
	last := page.authorities[token.limit-1]
	page.authorities = page.authorities[:token.limit]
	page.next = OutboxAuthorityPageToken{
		owner: token.owner, limit: token.limit,
		cursor: outboxAuthorityCursor{
			present: true, scopeKind: last.scopeKind, scopeID: last.scopeID,
			groupID: last.groupID, domain: last.domain,
		},
	}
	page.hasNext = true
	return page
}
