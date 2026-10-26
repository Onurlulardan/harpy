package iam

import (
	"errors"
	"fmt"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/google/uuid"
)

var (
	ErrPortalAccessAuthentication = errors.New("IAM portal access authentication failed")
	ErrPortalAccessForbidden      = errors.New("IAM portal access is forbidden")
	ErrPortalAccessUnavailable    = errors.New("IAM portal access is unavailable")
)

type PortalAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
}

type PortalAccessDecision struct{ binding portaluser.AccessBinding }

func newPortalAccessDecision(binding portaluser.AccessBinding) (PortalAccessDecision, error) {
	if !binding.Valid() {
		return PortalAccessDecision{}, fmt.Errorf("IAM portal access decision is invalid")
	}
	return PortalAccessDecision{binding: binding}, nil
}

func (value PortalAccessDecision) Valid() bool            { return value.binding.Valid() }
func (value PortalAccessDecision) PrincipalID() uuid.UUID { return value.binding.PrincipalID }
func (value PortalAccessDecision) TenantID() uuid.UUID    { return value.binding.TenantID }
func (value PortalAccessDecision) BindingID() uuid.UUID   { return value.binding.BindingID }
func (value PortalAccessDecision) TokenVersion() int64    { return value.binding.TokenVersion }
func (value PortalAccessDecision) AuthorizationVersion() int64 {
	return value.binding.AuthorizationVersion
}
func (value PortalAccessDecision) AuthorizationClock() int64 {
	return value.binding.AuthorizationClock
}
func (value PortalAccessDecision) BindingRevision() int64 { return value.binding.BindingRevision }
func (value PortalAccessDecision) SubjectType() string    { return value.binding.SubjectType }
func (value PortalAccessDecision) RoleNames() []string {
	return append([]string(nil), value.binding.RoleNames...)
}
func (value PortalAccessDecision) FormID() uuid.UUID        { return value.binding.FormID }
func (value PortalAccessDecision) FormVersionID() uuid.UUID { return value.binding.FormVersionID }
func (value PortalAccessDecision) SubjectRecordID() uuid.UUID {
	return value.binding.SubjectRecordID
}
func (value PortalAccessDecision) IdentifierFieldID() uuid.UUID {
	return value.binding.IdentifierFieldID
}
