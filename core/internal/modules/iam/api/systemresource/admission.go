package systemresource

const AuthorityGroup = "iam_system_resources"

type Admission struct {
	tenantID        ID
	actor           ActorEvidence
	groupGeneration int64
	validated       bool
}

func NewAdmission(
	tenantID ID,
	actor ActorEvidence,
	groupGeneration int64,
	iamGeneration int64,
	layoutGeneration int64,
) (Admission, error) {
	actorTenantID, actorErr := actor.TenantID()
	admission := Admission{
		tenantID: tenantID, actor: actor,
		groupGeneration: groupGeneration, validated: true,
	}
	if actorErr != nil || actorTenantID != tenantID || !tenantID.Valid() ||
		groupGeneration <= 0 || iamGeneration != groupGeneration ||
		layoutGeneration != groupGeneration {
		return Admission{}, ErrContract
	}
	return admission, nil
}

func (admission Admission) Validate() error {
	actorTenantID, actorErr := admission.actor.TenantID()
	if !admission.validated || !admission.tenantID.Valid() || actorErr != nil ||
		actorTenantID != admission.tenantID || admission.groupGeneration <= 0 {
		return ErrContract
	}
	return nil
}

func (admission Admission) Actor() (ActorEvidence, error) {
	if admission.Validate() != nil {
		return ActorEvidence{}, ErrContract
	}
	return admission.actor, nil
}

func (admission Admission) Identity() (ID, ID, int64, error) {
	if admission.Validate() != nil {
		return ID{}, ID{}, 0, ErrContract
	}
	actorID, err := admission.actor.ActorID()
	if err != nil {
		return ID{}, ID{}, 0, ErrContract
	}
	return actorID, admission.tenantID, admission.groupGeneration, nil
}

func (admission Admission) Generations() (map[string]int64, error) {
	if admission.Validate() != nil {
		return nil, ErrContract
	}
	return map[string]int64{
		"iam":                    admission.groupGeneration,
		"system_resource_layout": admission.groupGeneration,
	}, nil
}
