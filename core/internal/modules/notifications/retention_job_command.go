package notifications

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type retentionDatabaseJobCommand struct {
	commands *retentionapp.Commands
	buildErr error
}

func (command *retentionDatabaseJobCommand) Execute(ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.buildErr != nil || command.commands == nil || claim == nil {
		return platformasync.DatabaseJobOutcome{}, retentionapp.ErrConfiguration
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped {
		return platformasync.DatabaseJobOutcome{}, retentionapp.ErrClaimInvalid
	}
	outcome, err := command.commands.Purge(ctx, retentionapp.Claim{
		JobID: claim.JobID(), AttemptID: claim.AttemptID(), TenantID: tenantID,
		Kind: claim.Kind(), PayloadRevision: claim.PayloadRevision(),
		PayloadReference: claim.PayloadReference(), PayloadSHA256: claim.PayloadSHA256(),
	})
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"execute notification retention job: %w", err)
	}
	next := outcome.Successor
	return platformasync.ContinueDatabaseJob(platformasync.DatabaseJobSuccessorInput{
		JobID: next.JobID, RunAt: next.RunAt, PayloadReference: next.PayloadReference,
		PayloadSHA256: next.PayloadSHA256, DedupeKey: next.DedupeKey,
		Priority: next.Priority,
	})
}

var _ platformasync.DatabaseJobCommand = (*retentionDatabaseJobCommand)(nil)
