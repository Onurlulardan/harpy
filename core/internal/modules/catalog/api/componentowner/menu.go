package componentowner

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type MenuAuthorizer interface {
	AllowsMenuTarget(context.Context, platformtx.TenantScope,
		string, string) (bool, error)
}
