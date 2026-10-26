package iam

import queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"

func (runtime *IAMRESTQueryRuntime) newSystemResourceQuery(
	keys queryapp.CursorKeys,
) (*queryapp.Service, error) {
	if !runtime.Valid() {
		return nil, queryapp.ErrContract
	}
	return queryapp.NewService(
		runtime.users, runtime.roles, runtime.organizations,
		runtime.policy, runtime.execution, keys,
	)
}
