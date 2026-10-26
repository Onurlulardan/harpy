package policypostgres

import (
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

func NewAuthorizePolicyPort() (restauthorization.PolicyPort, error) {
	return NewPolicyService()
}
