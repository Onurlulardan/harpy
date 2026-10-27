package platformiam

type EvaluateQuery struct {
	ActorID                      ID
	Capability                   string
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	ExpectedPlatformClock        int64
}

func (query EvaluateQuery) Validate() error {
	if !query.ActorID.Valid() || !CapabilityKnown(query.Capability) ||
		query.ExpectedTokenVersion < 1 || query.ExpectedAuthorizationVersion < 1 ||
		query.ExpectedPlatformClock < 1 {
		return ErrInvalidQuery
	}
	return nil
}

type Decision struct {
	Allowed              bool
	Capability           string
	DeploymentMode       string
	SubjectID            ID
	PlatformClock        int64
	AuthorizationVersion int64
	Reason               string
}

func (decision Decision) Validate(query EvaluateQuery) error {
	validReason := decision.Reason == "granted" || decision.Reason == "not_granted"
	if decision.Capability != query.Capability ||
		(decision.DeploymentMode != "onpremise" && decision.DeploymentMode != "saas") ||
		(decision.Allowed && !decision.SubjectID.Valid()) || decision.PlatformClock < 1 ||
		decision.AuthorizationVersion < 1 || !validReason ||
		(decision.Allowed != (decision.Reason == "granted")) {
		return ErrStorage
	}
	return nil
}
