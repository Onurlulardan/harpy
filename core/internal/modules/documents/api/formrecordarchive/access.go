package formrecordarchive

type Access struct {
	actorID, tenantID                                      ID
	tokenVersion, authorizationVersion, authorizationClock int64
}

func NewAccess(actorID, tenantID ID, tokenVersion,
	authorizationVersion, authorizationClock int64,
) (Access, error) {
	access := Access{actorID: actorID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock}
	if !access.Valid() {
		return Access{}, ErrForbidden
	}
	return access, nil
}

func (access Access) Valid() bool {
	return access.actorID.Valid() && access.tenantID.Valid() &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0
}

func (access Access) ActorID() ID                 { return access.actorID }
func (access Access) TenantID() ID                { return access.tenantID }
func (access Access) TokenVersion() int64         { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 { return access.authorizationVersion }
func (access Access) AuthorizationClock() int64   { return access.authorizationClock }
