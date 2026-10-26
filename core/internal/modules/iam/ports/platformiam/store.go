package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

type Store[S any] interface {
	Grant(context.Context, S, platformdomain.OperatorGrant) (platformdomain.MutationResult, error)
	Revoke(context.Context, S, platformdomain.OperatorRevoke) (platformdomain.MutationResult, error)
	Authorize(context.Context, S, platformdomain.EvaluateQuery) (platformdomain.Decision, error)
	ReadSessionContext(context.Context, S, platformdomain.SessionContextQuery) (platformdomain.SessionContext, error)
}
