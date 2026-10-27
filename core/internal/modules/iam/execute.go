package iam

import (
	"context"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type maintenanceDatabaseJobCommand struct {
	commands *maintenanceapp.Commands
	buildErr error
}

func (command *maintenanceDatabaseJobCommand) Execute(
	ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.buildErr != nil || command.commands == nil ||
		claim == nil {
		return platformasync.DatabaseJobOutcome{}, maintenanceapp.ErrCommandInvalid
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped {
		return platformasync.DatabaseJobOutcome{}, maintenanceapp.ErrClaimInvalid
	}
	outcome, err := command.commands.Execute(ctx, maintenanceapp.Claim{
		JobID: claim.JobID(), AttemptID: claim.AttemptID(), TenantID: tenantID,
		Kind:             maintenanceapp.Kind(claim.Kind()),
		PayloadRevision:  claim.PayloadRevision(),
		PayloadReference: claim.PayloadReference(),
		PayloadSHA256:    claim.PayloadSHA256(),
	})
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf("execute IAM maintenance job: %w", err)
	}
	if outcome.Successor == nil {
		return platformasync.CompleteDatabaseJob(), nil
	}
	next := outcome.Successor
	return platformasync.ContinueDatabaseJob(platformasync.DatabaseJobSuccessorInput{
		JobID: next.JobID, RunAt: next.RunAt,
		PayloadReference: next.PayloadReference,
		PayloadSHA256:    next.PayloadSHA256, DedupeKey: next.DedupeKey,
		Priority: next.Priority,
	})
}
