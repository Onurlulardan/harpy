package provider

import (
	"context"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func (storage *finalizerStore) FinalizeCommand(
	ctx context.Context,
	mutation providerapp.FinalizeCommandMutation,
) (providerapp.Finalization, error) {
	if !storage.usable(ctx) || mutation.TenantID != storage.tenantID {
		return providerapp.Finalization{}, providerapp.ErrCommandRequest
	}
	result, err := scanFinalization(storage.database.queryRow(
		ctx, finalizeCommandSQL, mutation.TenantID, mutation.JobID,
		mutation.AttemptID, mutation.LeaseGeneration, mutation.PayloadRevision,
		mutation.PayloadReference, mutation.PayloadSHA256, mutation.CommandID,
		mutation.NextFinalizerJobID, mutation.ReconciliationJobID,
		mutation.ObservedAt,
	))
	if err != nil {
		return providerapp.Finalization{}, fmt.Errorf(
			"persist billing provider finalization: %w",
			classifyCommandError(err),
		)
	}
	return result, nil
}
