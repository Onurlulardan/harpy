package deploymentpolicy

import "errors"

var errPolicyUnavailable = errors.New(
	"tenancy deployment policy storage is unavailable",
)
