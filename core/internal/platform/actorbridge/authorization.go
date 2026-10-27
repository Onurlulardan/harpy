package actorbridge

import "github.com/google/uuid"

type Authorization struct {
	tenantID, subjectID, permissionID  uuid.UUID
	tokenVersion, authorizationVersion int64
	authorizationClock                 int64
}

func NewAuthorization(
	tenantID uuid.UUID,
	subjectID uuid.UUID,
	permissionID uuid.UUID,
	tokenVersion int64,
	authorizationVersion int64,
	authorizationClock int64,
) (Authorization, error) {
	authorization := Authorization{tenantID: tenantID, subjectID: subjectID,
		permissionID: permissionID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion,
		authorizationClock:   authorizationClock}
	if authorization.Validate() != nil {
		return Authorization{}, NewFault(FaultInvalid)
	}
	return authorization, nil
}

func (authorization Authorization) Validate() error {
	if authorization.tenantID == uuid.Nil || authorization.subjectID == uuid.Nil ||
		authorization.permissionID == uuid.Nil || authorization.tokenVersion <= 0 ||
		authorization.authorizationVersion <= 0 || authorization.authorizationClock < 0 {
		return NewFault(FaultInvalid)
	}
	return nil
}

func (authorization Authorization) TenantID() uuid.UUID  { return authorization.tenantID }
func (authorization Authorization) SubjectID() uuid.UUID { return authorization.subjectID }
func (authorization Authorization) PermissionID() uuid.UUID {
	return authorization.permissionID
}
func (authorization Authorization) TokenVersion() int64 { return authorization.tokenVersion }
func (authorization Authorization) AuthorizationVersion() int64 {
	return authorization.authorizationVersion
}
func (authorization Authorization) AuthorizationClock() int64 {
	return authorization.authorizationClock
}
