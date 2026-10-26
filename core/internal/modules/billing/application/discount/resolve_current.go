package discount

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (queries *Queries) ResolveCurrent(
	ctx context.Context,
	scope platformtx.TenantScope,
	code string,
) (Eligibility, error) {
	parsed, err := ParseCode(code)
	if err != nil {
		return Eligibility{}, ErrQueryInvariant
	}
	return queries.find(ctx, scope, parsed, uuid.Nil)
}
