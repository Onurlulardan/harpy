package systemtransfer

import (
	"context"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *exportWorkerBound) PersistOutboundFailure(
	ctx context.Context,
	failure app.ExportFailure,
) error {
	if !store.usable(ctx, failure.Attempt) || !validExportFailure(failure) {
		return app.ErrExportWorkerInvalid
	}
	schema, err := store.schema(ctx)
	if err != nil {
		return err
	}
	arguments := append(outboundAttemptArguments(schema, failure.Attempt),
		string(failure.State), failure.Code, failure.At)
	var replayed bool
	err = store.database.QueryRow(ctx, outboundFinishSQL,
		arguments...).Scan(&replayed)
	if err != nil {
		return mapOutboundError(err, app.ErrExportStateConflict)
	}
	return nil
}

func validExportFailure(failure app.ExportFailure) bool {
	if failure.At.IsZero() || failure.At.Location() != time.UTC ||
		failure.At.Nanosecond()%int(time.Microsecond) != 0 {
		return false
	}
	if failure.State == app.ExportFinalCancelled {
		return failure.Code == app.ExportCodeCancelled
	}
	if failure.State != app.ExportFinalFailed {
		return false
	}
	switch failure.Code {
	case app.ExportCodeAuthorizationRevoked, app.ExportCodeLimitExceeded,
		app.ExportCodeQueryInvalid, app.ExportCodeCatalogPinStale,
		app.ExportCodeArtifactTampered, app.ExportCodeStateConflict,
		app.ExportCodeDependencyUnavailable:
		return true
	default:
		return false
	}
}
