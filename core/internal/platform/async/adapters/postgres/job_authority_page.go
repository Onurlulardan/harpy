package postgres

import "github.com/google/uuid"

const maximumJobAuthorityPageSize = 256

type jobAuthorityCursor struct {
	present   bool
	scopeKind string
	scopeID   uuid.UUID
	groupID   string
	domain    string
}

type JobAuthorityPageToken struct {
	owner  *workerIdentity
	limit  int
	cursor jobAuthorityCursor
}

type JobAuthorityPage struct {
	authorities []JobAuthority
	next        JobAuthorityPageToken
	hasNext     bool
}

func (repository *workerRepository) BeginJobAuthorityPage(
	limit int,
) (JobAuthorityPageToken, error) {
	if !repository.valid() {
		return JobAuthorityPageToken{}, ErrWorkerStoreUnavailable
	}
	if limit < 1 || limit > maximumJobAuthorityPageSize {
		return JobAuthorityPageToken{}, ErrJobAuthorityPageLimit
	}
	return JobAuthorityPageToken{owner: repository.identity, limit: limit}, nil
}

func (token JobAuthorityPageToken) valid(repository *workerRepository) bool {
	if !repository.valid() || token.owner != repository.identity ||
		token.limit < 1 || token.limit > maximumJobAuthorityPageSize {
		return false
	}
	if !token.cursor.present {
		return token.cursor == (jobAuthorityCursor{})
	}
	return (token.cursor.scopeKind == "tenant" || token.cursor.scopeKind == "global") &&
		token.cursor.scopeID != uuid.Nil && token.cursor.groupID != "" &&
		token.cursor.domain != ""
}

func (page JobAuthorityPage) Authorities() []JobAuthority {
	return append([]JobAuthority(nil), page.authorities...)
}

func (page JobAuthorityPage) Next() (JobAuthorityPageToken, bool) {
	return page.next, page.hasNext
}

func makeJobAuthorityPage(
	token JobAuthorityPageToken,
	authorities []JobAuthority,
) JobAuthorityPage {
	page := JobAuthorityPage{authorities: append([]JobAuthority(nil), authorities...)}
	if len(page.authorities) <= token.limit {
		return page
	}
	last := page.authorities[token.limit-1]
	page.authorities = page.authorities[:token.limit]
	page.next = JobAuthorityPageToken{
		owner: token.owner, limit: token.limit,
		cursor: jobAuthorityCursor{
			present: true, scopeKind: last.scopeKind, scopeID: last.scopeID,
			groupID: last.groupID, domain: last.domain,
		},
	}
	page.hasNext = true
	return page
}
