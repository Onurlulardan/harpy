package systemresource

type ActorEvidence struct {
	tenantID             ID
	actorID              ID
	tokenVersion         int64
	authorizationVersion int64
	validated            bool
}

func NewActorEvidence(
	tenantID ID,
	actorID ID,
	tokenVersion int64,
	authorizationVersion int64,
) (ActorEvidence, error) {
	evidence := ActorEvidence{
		tenantID: tenantID, actorID: actorID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, validated: true,
	}
	if evidence.Validate() != nil {
		return ActorEvidence{}, ErrContract
	}
	return evidence, nil
}

func (evidence ActorEvidence) Validate() error {
	if !evidence.validated || !evidence.tenantID.Valid() ||
		!evidence.actorID.Valid() || evidence.tokenVersion < 1 ||
		evidence.authorizationVersion < 1 {
		return ErrContract
	}
	return nil
}

func (evidence ActorEvidence) TenantID() (ID, error) {
	if evidence.Validate() != nil {
		return ID{}, ErrContract
	}
	return evidence.tenantID, nil
}

func (evidence ActorEvidence) ActorID() (ID, error) {
	if evidence.Validate() != nil {
		return ID{}, ErrContract
	}
	return evidence.actorID, nil
}

func (evidence ActorEvidence) Versions() (int64, int64, error) {
	if evidence.Validate() != nil {
		return 0, 0, ErrContract
	}
	return evidence.tokenVersion, evidence.authorizationVersion, nil
}
