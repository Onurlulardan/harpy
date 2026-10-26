package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *OutboundAttemptStore) LoadExportAttempt(
	ctx context.Context,
	claim app.ExportJobClaim,
) (app.ExportAttempt, error) {
	var attempt app.ExportAttempt
	err := store.inExportTenant(ctx, claim, func(scoped context.Context,
		bound *exportAttemptBound) error {
		schema, err := bound.database.CurrentSchema(scoped)
		if err != nil {
			return app.ErrExportUnavailable
		}
		var raw outboundAttemptRow
		err = bound.database.QueryRow(scoped, outboundLoadSQL, schema,
			claim.TenantID, claim.ExportID, claim.JobID, string(claim.Resource),
		).Scan(raw.targets()...)
		if err != nil {
			return mapOutboundError(err, app.ErrExportStateConflict)
		}
		attempt, err = raw.decode(claim)
		return err
	})
	return attempt, err
}
