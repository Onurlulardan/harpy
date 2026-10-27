package analytics

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/retention"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type analyticsRetentionCommand struct {
	commands *retentionapp.Commands
	buildErr error
}

func (command *analyticsRetentionCommand) Execute(
	ctx context.Context, claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.buildErr != nil || command.commands == nil || claim == nil {
		return platformasync.DatabaseJobOutcome{}, retentionapp.ErrConfiguration
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped {
		return platformasync.DatabaseJobOutcome{}, retentionapp.ErrClaimInvalid
	}
	outcome, err := command.commands.Execute(ctx, retentionapp.Claim{
		JobID: claim.JobID(), AttemptID: claim.AttemptID(), TenantID: tenantID,
		Kind: claim.Kind(), PayloadRevision: claim.PayloadRevision(),
		PayloadReference: claim.PayloadReference(), PayloadSHA256: claim.PayloadSHA256(),
	})
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"execute Analytics retention job: %w", err)
	}
	next := outcome.Successor
	return platformasync.ContinueDatabaseJob(platformasync.DatabaseJobSuccessorInput{
		JobID: next.JobID, RunAt: next.RunAt,
		PayloadReference: next.PayloadReference, PayloadSHA256: next.PayloadSHA256,
		DedupeKey: next.DedupeKey, Priority: next.Priority,
	})
}

var _ platformasync.DatabaseJobCommand = (*analyticsRetentionCommand)(nil)
