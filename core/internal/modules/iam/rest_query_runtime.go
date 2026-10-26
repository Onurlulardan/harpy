package iam

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	executionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/execution"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/organization"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/policy"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/role"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/user"
)

// IAMRESTQueryRuntime is IAM's final owner surface for catalog, policy and
// snapshot execution. Composition never constructs its application graph.
type IAMRESTQueryRuntime struct {
	users         *userapp.Service
	roles         *roleapp.Service
	organizations *organizationapp.Service
	policy        *policyapp.Service
	execution     *executionapp.Executor
}

func NewIAMRESTQueryRuntime() (*IAMRESTQueryRuntime, error) {
	users, roles, organizations, err := buildIAMRESTQueryCatalogs()
	if err != nil {
		return nil, err
	}
	policy, err := buildIAMRESTQueryPolicy()
	if err != nil {
		return nil, err
	}
	execution, err := buildIAMRESTQueryExecution()
	if err != nil {
		return nil, err
	}
	return &IAMRESTQueryRuntime{
		users: users, roles: roles, organizations: organizations,
		policy: policy, execution: execution,
	}, nil
}

func (runtime *IAMRESTQueryRuntime) Valid() bool {
	return runtime != nil && runtime.users != nil && runtime.roles != nil &&
		runtime.organizations != nil && runtime.policy != nil && runtime.execution != nil
}

func (runtime *IAMRESTQueryRuntime) Users() *userapp.Service {
	if !runtime.Valid() {
		return nil
	}
	return runtime.users
}

func (runtime *IAMRESTQueryRuntime) Roles() *roleapp.Service {
	if !runtime.Valid() {
		return nil
	}
	return runtime.roles
}

func (runtime *IAMRESTQueryRuntime) Organizations() *organizationapp.Service {
	if !runtime.Valid() {
		return nil
	}
	return runtime.organizations
}

func (runtime *IAMRESTQueryRuntime) Policy() *policyapp.Service {
	if !runtime.Valid() {
		return nil
	}
	return runtime.policy
}

func (runtime *IAMRESTQueryRuntime) Execution() *executionapp.Executor {
	if !runtime.Valid() {
		return nil
	}
	return runtime.execution
}

func (*IAMRESTQueryRuntime) Owner() string { return catalog.Owner }
