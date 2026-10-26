package postgres

import "github.com/google/uuid"

type JobAuthority struct {
	owner           *workerIdentity
	mode            jobAuthorityMode
	scopeKind       string
	scopeID         uuid.UUID
	groupID         string
	groupGeneration int64
	domain          string
	generation      int64
}

func newJobAuthority(
	owner *workerIdentity,
	scopeKind string,
	scopeID uuid.UUID,
	groupID string,
	groupGeneration int64,
	domain string,
	generation int64,
) (JobAuthority, error) {
	return newJobAuthorityWithMode(owner, jobAuthorityModeV2, scopeKind, scopeID,
		groupID, groupGeneration, domain, generation)
}

func newJobAuthorityWithMode(
	owner *workerIdentity,
	mode jobAuthorityMode,
	scopeKind string,
	scopeID uuid.UUID,
	groupID string,
	groupGeneration int64,
	domain string,
	generation int64,
) (JobAuthority, error) {
	authority := JobAuthority{
		owner: owner, mode: mode, scopeKind: scopeKind, scopeID: scopeID,
		groupID: groupID, groupGeneration: groupGeneration,
		domain: domain, generation: generation,
	}
	if !authority.valid(owner) {
		return JobAuthority{}, ErrJobAuthorityInvalid
	}
	return authority, nil
}

func (authority JobAuthority) valid(owner *workerIdentity) bool {
	return owner != nil && owner.marker == 1 && authority.owner == owner &&
		authority.mode.valid() &&
		(authority.scopeKind == "tenant" || authority.scopeKind == "global") &&
		authority.scopeID != uuid.Nil && authority.groupID != "" &&
		authority.groupGeneration > 0 && authority.domain != "" &&
		authority.generation > 0
}

func (authority JobAuthority) TenantID() (uuid.UUID, bool) {
	if !authority.valid(authority.owner) || authority.scopeKind != "tenant" {
		return uuid.Nil, false
	}
	return authority.scopeID, true
}
