package deploymentpolicy

import (
	"context"

	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/deploymentpolicy"
)

type policyInitializer struct{ queryRow QueryRow }

func NewInitializer(queryRow QueryRow) (policyapp.InitializationStore, error) {
	if queryRow == nil {
		return nil, errPolicyUnavailable
	}
	return &policyInitializer{queryRow: queryRow}, nil
}

func (initializer *policyInitializer) usable(ctx context.Context) bool {
	return initializer != nil && initializer.queryRow != nil &&
		ctx != nil && ctx.Err() == nil
}

var _ policyapp.InitializationStore = (*policyInitializer)(nil)
