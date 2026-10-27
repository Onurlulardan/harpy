package platformiam

type SessionContextQuery struct {
	ActorID                      ID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (query SessionContextQuery) Validate() error {
	if !query.ActorID.Valid() || query.ExpectedTokenVersion < 1 ||
		query.ExpectedAuthorizationVersion < 1 {
		return ErrInvalidQuery
	}
	return nil
}

type SessionContext struct {
	DeploymentMode       string
	TenantPurpose        string
	SubjectID            ID
	PlatformClock        int64
	AuthorizationVersion int64
}

func (context SessionContext) Validate() error {
	validMode := context.DeploymentMode == "onpremise" || context.DeploymentMode == "saas"
	primary := context.TenantPurpose == "onpremise_primary" ||
		context.TenantPurpose == "saas_control"
	customer := context.TenantPurpose == "saas_customer"
	if !validMode || context.AuthorizationVersion < 1 ||
		(primary && context.PlatformClock < 1) ||
		(customer && (context.DeploymentMode != "saas" || context.PlatformClock != 0 ||
			context.SubjectID.Valid())) || (!primary && !customer) {
		return ErrStorage
	}
	return nil
}
