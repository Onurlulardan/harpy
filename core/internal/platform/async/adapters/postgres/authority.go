package postgres

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type authorityStamp struct {
	scopeID         uuid.UUID
	groupID         string
	groupGeneration int64
	domain          string
	generation      int64
}

func newAuthorityStamp(authority transaction.OwnerAuthority) (authorityStamp, error) {
	stamp := authorityStamp{
		scopeID: authority.ScopeID(), groupID: authority.GroupID(),
		groupGeneration: authority.GroupGeneration(), domain: authority.Domain(),
		generation: authority.Generation(),
	}
	if authority.ScopeKind() != "tenant" || stamp.scopeID == uuid.Nil ||
		stamp.groupID == "" || stamp.groupGeneration < 1 || stamp.domain == "" ||
		stamp.generation < 1 || authority.Origin() != transaction.WriterOriginV2 {
		return authorityStamp{}, fmt.Errorf("capture async owner authority: %w", ErrAuthorityMismatch)
	}
	return stamp, nil
}

func (stamp authorityStamp) matchesJob(value jobsv1.Enqueue) bool {
	authority := value.Authority
	return value.TenantID != nil && *value.TenantID == stamp.scopeID.String() &&
		authority.ScopeKind == jobsv1.ScopeTenant &&
		authority.ScopeID == stamp.scopeID.String() && authority.GroupID == stamp.groupID &&
		authority.GroupGeneration == stamp.groupGeneration && authority.Domain == stamp.domain &&
		authority.CutoverGeneration == stamp.generation && authority.WriterOrigin == jobsv1.WriterV2
}

func (stamp authorityStamp) matchesEvent(value eventsv1.Envelope) bool {
	authority := value.ProducerAuthority
	return value.TenantID != nil && *value.TenantID == stamp.scopeID.String() &&
		authority.ScopeKind == eventsv1.ScopeTenant &&
		authority.ScopeID == stamp.scopeID.String() && authority.GroupID == stamp.groupID &&
		authority.GroupGeneration == stamp.groupGeneration && authority.Domain == stamp.domain &&
		authority.CutoverGeneration == stamp.generation && authority.WriterOrigin == eventsv1.WriterV2
}

func (stamp authorityStamp) matchesRetentionRelease(value jobsv1.RetentionRelease) bool {
	authority := value.Authority
	return value.TenantID != nil && *value.TenantID == stamp.scopeID.String() &&
		authority.ScopeKind == jobsv1.ScopeTenant &&
		authority.ScopeID == stamp.scopeID.String() && authority.GroupID == stamp.groupID &&
		authority.GroupGeneration == stamp.groupGeneration && authority.Domain == stamp.domain &&
		authority.CutoverGeneration == stamp.generation && authority.WriterOrigin == jobsv1.WriterV2
}

func (stamp authorityStamp) matchesClaim(claim JobClaim) bool {
	return claim.valid(claim.authority.owner) && claim.tenantScoped &&
		claim.tenantID == stamp.scopeID && claim.authority.scopeKind == "tenant" &&
		claim.authority.scopeID == stamp.scopeID &&
		claim.authority.groupID == stamp.groupID &&
		claim.authority.groupGeneration == stamp.groupGeneration &&
		claim.authority.domain == stamp.domain &&
		claim.authority.generation == stamp.generation
}
