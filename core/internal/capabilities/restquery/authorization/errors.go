package authorization

import "errors"

var errInvalidRequirementIdentity = errors.New(
	"REST Query authorization requirement identity is invalid",
)

var errInvalidPolicyRequestIdentity = errors.New(
	"REST Query policy request identity is invalid",
)

var errPolicyDecisionContract = errors.New(
	"REST Query policy decision contract is invalid",
)

var errPolicyEvaluationContract = errors.New(
	"REST Query policy evaluation contract is invalid",
)

var errSnapshotOperationAdmission = errors.New(
	"REST Query snapshot operation admission is invalid",
)
