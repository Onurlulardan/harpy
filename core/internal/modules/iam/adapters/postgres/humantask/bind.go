package humantask

import (
	"context"
	"errors"
	"strings"

	humantaskapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/humantask"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errCandidateRead = errors.New("IAM workflow human task candidate evidence failed")

type tenantScope = platformtx.TenantScope

type boundReader struct {
	database platformpostgres.BoundExecutor
}

type reader struct {
	binder platformpostgres.OwnerBinder[*boundReader]
	valid  bool
}

func NewCandidateReader() (humantaskapp.EvidencePort, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"workflow",
		func(database platformpostgres.BoundExecutor) *boundReader {
			return &boundReader{database: database}
		},
	)
	if err != nil {
		return nil, errCandidateRead
	}
	return &reader{binder: binder, valid: true}, nil
}

var _ humantaskapp.CandidatePort = (*reader)(nil)
var _ humantaskapp.AssignmentPort = (*reader)(nil)

func (reader *reader) bind(scope tenantScope) (*boundReader, error) {
	if reader == nil || !reader.valid {
		return nil, errCandidateRead
	}
	bound, err := platformtx.Bind(scope, reader.binder)
	if err != nil || bound == nil {
		return nil, errCandidateRead
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		strings.HasPrefix(scope.OperationID(), "workflow.human_task.")
}
