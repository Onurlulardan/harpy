package v1adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	"github.com/google/uuid"
)

func (store *Store) BeginAndLock(ctx context.Context, plan app.Plan) (app.Run, error) {
	if !store.usable(ctx) || plan.Validate() != nil || plan.TenantID != store.tenantID {
		return app.Run{}, app.ErrInvalid
	}
	if err := store.execute(ctx, insertRunSQL, plan.RunID,
		plan.SourceGeneration, plan.SourceManifestDigest[:]); err != nil {
		return app.Run{}, app.ErrUnavailable
	}
	var run app.Run
	var lastSourceID *uuid.UUID
	var sourceChecksum, targetChecksum []byte
	err := store.queryRow(ctx, lockRunSQL, plan.SourceGeneration,
		plan.SourceManifestDigest[:]).Scan(&run.ID, &run.State, &lastSourceID,
		&run.Checkpoint.AdoptedCount, &sourceChecksum, &targetChecksum, &run.Unresolved)
	if err != nil {
		return app.Run{}, app.ErrBlocked
	}
	if len(sourceChecksum) != 0 && len(sourceChecksum) != 32 ||
		len(targetChecksum) != 0 && len(targetChecksum) != 32 {
		return app.Run{}, app.ErrUnavailable
	}
	copy(run.Checkpoint.SourceChecksum[:], sourceChecksum)
	copy(run.Checkpoint.TargetChecksum[:], targetChecksum)
	run.Checkpoint.LastSourceID = lastSourceID
	if run.State != "planned" && run.State != "applying" && run.State != "verified" {
		return app.Run{}, app.ErrBlocked
	}
	if run.State == "planned" {
		if err = store.execute(ctx, beginApplyingSQL, run.ID); err != nil {
			return app.Run{}, app.ErrUnavailable
		}
		run.State = "applying"
	}
	return run, nil
}
