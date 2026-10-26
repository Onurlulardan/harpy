package securityepoch

import (
	"context"
	_ "embed"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/securityepoch"
)

//go:embed sql/tenant/assert.sql
var assertSQL string

func (storage *store) Assert(
	ctx context.Context, scope tenantScope, epoch [32]byte,
) error {
	bound, err := storage.bind(scope)
	if err != nil {
		return app.ErrRejected
	}
	var asserted bool
	if err = bound.database.QueryRow(ctx, assertSQL, epoch[:]).Scan(&asserted); err != nil || !asserted {
		return app.ErrRejected
	}
	return nil
}
