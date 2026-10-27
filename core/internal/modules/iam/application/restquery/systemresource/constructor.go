package systemresourcequery

import (
	"fmt"

	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	executionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/execution"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/organization"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/policy"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/role"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/user"
)

type CursorKeys = queryruntime.CursorKeys

func NewCursorKeys(
	currentID string, current [32]byte,
	previousID string, previous [32]byte,
) (CursorKeys, error) {
	return queryruntime.NewCursorKeys(currentID, current, previousID, previous)
}

type Service struct {
	engine   *queryruntime.Engine
	catalogs catalogs
}

func NewService(
	users *userapp.Service,
	roles *roleapp.Service,
	organizations *organizationapp.Service,
	policy *policyapp.Service,
	execution *executionapp.Executor,
	keys CursorKeys,
) (*Service, error) {
	if users == nil || roles == nil || organizations == nil ||
		policy == nil || execution == nil {
		return nil, fmt.Errorf("IAM system-resource query dependencies are required")
	}
	provider := catalogs{users: users, roles: roles, organizations: organizations}
	engine, err := queryruntime.New("iam", provider, policy, execution, keys)
	if err != nil {
		return nil, err
	}
	return &Service{engine: engine, catalogs: provider}, nil
}
