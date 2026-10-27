package provider

import (
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func scanCommandSnapshot(row databaseRow) (providerapp.CommandSnapshot, error) {
	var result providerapp.CommandSnapshot
	var mappingID *uuid.UUID
	var kind, state string
	var digest []byte
	err := row.Scan(
		&result.CommandID, &result.FinalizerJobID, &result.TenantID,
		&result.ActorID, &result.ProviderAccountID, &mappingID,
		&result.SubscriptionID, &result.Provider, &kind, &result.EffectKey,
		&digest, &state, &result.RequestedAt, &result.Revision, &result.Replay,
	)
	if err != nil {
		return providerapp.CommandSnapshot{}, err
	}
	if len(digest) != 32 {
		return providerapp.CommandSnapshot{}, fmt.Errorf(
			"billing provider command digest is invalid",
		)
	}
	copy(result.RequestSHA256[:], digest)
	result.ProviderObjectMappingID = mappingID
	result.Kind = providerapp.CommandKind(kind)
	result.State = providerapp.CommandState(state)
	return result, nil
}
