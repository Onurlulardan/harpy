package iam

import (
	"fmt"
	"time"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// NewWorkflowInitiatorRuntime builds the tenant-scope workflow start verifier.
// It does not load API bearer or CSRF secrets.
func NewWorkflowInitiatorRuntime(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
) (*WorkflowAccessRuntime, error) {
	if pool == nil || coordinator == nil {
		return nil, fmt.Errorf("IAM Workflow initiator dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM Workflow initiator authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM Workflow initiator policy: %w", err)
	}
	return &WorkflowAccessRuntime{shared: &FormsRecordAccessRuntime{
		coordinator: coordinator, authorizer: authorizer, clock: time.Now,
	}}, nil
}
