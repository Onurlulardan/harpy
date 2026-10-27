package workhttp

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAuthentication = errors.New("WORK_HTTP_AUTHENTICATION_FAILED")
	ErrForbidden      = errors.New("WORK_HTTP_FORBIDDEN")
	ErrInvalid        = errors.New("WORK_HTTP_INVALID")
	ErrNotFound       = errors.New("WORK_HTTP_NOT_FOUND")
	ErrConflict       = errors.New("WORK_HTTP_CONFLICT")
	ErrUnavailable    = errors.New("WORK_HTTP_UNAVAILABLE")
)

type AccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	ConsumerIDValues    []string
	BuildIDValues       []string
	Authority           string
	Operation           string
	Compatibility       bool
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (Access, error)
}

type Access struct {
	principalID, tenantID              uuid.UUID
	tokenVersion, authorizationVersion int64
	authority, operation               string
	consumerID, buildID                string
	compatibility                      bool
	protectedAdmin                     bool
}

func NewAccessWithEvidence(
	principalID, tenantID uuid.UUID, tokenVersion, authorizationVersion int64,
	authority, operation, consumerID, buildID string, compatibility, protectedAdmin bool,
) (Access, error) {
	value, err := NewAccess(
		principalID, tenantID, tokenVersion, authorizationVersion,
		authority, operation, consumerID, buildID, compatibility,
	)
	if err != nil {
		return Access{}, err
	}
	value.protectedAdmin = protectedAdmin
	return value, nil
}

func NewAccess(principalID, tenantID uuid.UUID, tokenVersion, authorizationVersion int64,
	authority, operation, consumerID, buildID string, compatibility bool,
) (Access, error) {
	value := Access{principalID: principalID, tenantID: tenantID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, authority: authority, operation: operation,
		consumerID: consumerID, buildID: buildID, compatibility: compatibility}
	if !value.ValidFor(authority, operation, compatibility) {
		return Access{}, ErrInvalid
	}
	return value, nil
}

func (value Access) ValidFor(authority, operation string, compatibility bool) bool {
	if value.principalID == uuid.Nil || value.tenantID == uuid.Nil ||
		value.tokenVersion < 1 || value.authorizationVersion < 1 ||
		value.authority != authority || value.operation != operation ||
		value.compatibility != compatibility || authority == "" || operation == "" {
		return false
	}
	return true
}

func (value Access) PrincipalID() uuid.UUID { return value.principalID }
func (value Access) TenantID() uuid.UUID    { return value.tenantID }
func (value Access) TokenVersion() int64    { return value.tokenVersion }
func (value Access) AuthorizationVersion() int64 {
	return value.authorizationVersion
}
func (value Access) Authority() string    { return value.authority }
func (value Access) ConsumerID() string   { return value.consumerID }
func (value Access) BuildID() string      { return value.buildID }
func (value Access) ProtectedAdmin() bool { return value.protectedAdmin }
