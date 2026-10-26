package iam

import (
	"strings"

	"github.com/google/uuid"
)

// PortalActionGrant is server-minted evidence for one verified portal action.
// Its fields are intentionally private so HTTP input cannot construct a grant.
type PortalActionGrant struct {
	principalID, tenantID, bindingID, permissionID uuid.UUID
	tokenVersion, authorizationVersion, clock      int64
	resource, action                               string
}

func newPortalActionGrant(
	access PortalAccessDecision,
	permissionID uuid.UUID,
	authorizationVersion, clock int64,
	resource, action string,
) (PortalActionGrant, error) {
	grant := PortalActionGrant{
		principalID: access.PrincipalID(), tenantID: access.TenantID(),
		bindingID: access.BindingID(), permissionID: permissionID,
		tokenVersion: access.TokenVersion(), authorizationVersion: authorizationVersion,
		clock: clock, resource: strings.TrimSpace(resource), action: strings.TrimSpace(action),
	}
	if !grant.ValidFor(access, resource, action) {
		return PortalActionGrant{}, ErrPortalAccessUnavailable
	}
	return grant, nil
}

func (grant PortalActionGrant) ValidFor(
	access PortalAccessDecision, resource, action string,
) bool {
	return access.Valid() && grant.principalID == access.PrincipalID() &&
		grant.tenantID == access.TenantID() && grant.bindingID == access.BindingID() &&
		grant.permissionID != uuid.Nil && grant.tokenVersion == access.TokenVersion() &&
		grant.authorizationVersion == access.AuthorizationVersion() &&
		grant.clock >= access.AuthorizationClock() && grant.clock > 0 &&
		grant.resource == strings.TrimSpace(resource) && grant.resource != "" &&
		grant.action == strings.TrimSpace(action) && grant.action != ""
}

func (grant PortalActionGrant) PrincipalID() uuid.UUID      { return grant.principalID }
func (grant PortalActionGrant) TenantID() uuid.UUID         { return grant.tenantID }
func (grant PortalActionGrant) BindingID() uuid.UUID        { return grant.bindingID }
func (grant PortalActionGrant) PermissionID() uuid.UUID     { return grant.permissionID }
func (grant PortalActionGrant) TokenVersion() int64         { return grant.tokenVersion }
func (grant PortalActionGrant) AuthorizationVersion() int64 { return grant.authorizationVersion }
func (grant PortalActionGrant) AuthorizationClock() int64   { return grant.clock }
func (grant PortalActionGrant) Resource() string            { return grant.resource }
func (grant PortalActionGrant) Action() string              { return grant.action }
