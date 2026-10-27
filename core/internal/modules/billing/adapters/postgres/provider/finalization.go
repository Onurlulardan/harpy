package provider

import (
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func scanFinalization(row databaseRow) (providerapp.Finalization, error) {
	var result providerapp.Finalization
	var decision, state string
	var providerReference, errorCode nullableString
	err := row.Scan(
		&decision, &result.CommandID, &result.ActorID, &result.Provider,
		&result.ActiveJobID, &result.FinalizerJobID, &result.RequestSHA256,
		&state, &result.Revision, &result.OccurredAt, &providerReference,
		&errorCode,
	)
	if err != nil {
		return providerapp.Finalization{}, err
	}
	result.Decision = providerapp.FinalizeDecision(decision)
	result.State = providerapp.CommandState(state)
	if providerReference.valid {
		result.ProviderReference = providerReference.value
	}
	if errorCode.valid {
		result.ErrorCode = errorCode.value
	}
	return result, nil
}
