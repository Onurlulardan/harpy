package installationdelivery

import (
	"context"
	"time"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

type credentialStore struct {
	value Delivery
}

func (store *credentialStore) Stage(
	_ context.Context,
	command StageCommand,
) (Delivery, error) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	store.value = Delivery{
		OperationID: command.OperationID, TenantID: uuid.New(),
		FirstOwnerUserID: uuid.New(), CurrentAttemptID: command.AttemptID,
		CurrentCredentialID:      command.CredentialID,
		CurrentCredentialVersion: command.CredentialVersion,
		Revision:                 1, State: StateReady, DatabaseNow: now,
	}
	return store.value, nil
}

func (store *credentialStore) Read(
	context.Context,
	ReadQuery,
) (Delivery, error) {
	return store.value, nil
}

func (store *credentialStore) Claim(
	_ context.Context,
	command ClaimCommand,
) (Delivery, error) {
	now := store.value.DatabaseNow.Add(time.Second)
	until := now.Add(time.Duration(command.LeaseSeconds) * time.Second)
	store.value.State, store.value.Revision = StateClaimed, 2
	store.value.DatabaseNow = now
	store.value.ClaimAdapterID, store.value.ClaimLeaseUntil = command.AdapterID, &until
	return store.value, nil
}

func (store *credentialStore) Acknowledge(
	_ context.Context,
	command AcknowledgeCommand,
) (Delivery, error) {
	now := store.value.DatabaseNow.Add(time.Second)
	store.value.State, store.value.Revision = StateAcknowledged, 3
	store.value.DatabaseNow, store.value.AcknowledgedAt = now, &now
	store.value.ClaimAdapterID = command.AdapterID
	return store.value, nil
}

func (store *credentialStore) RecordUnknownOutcome(
	_ context.Context,
	_ UnknownCommand,
) (Transition, error) {
	now := store.value.DatabaseNow.Add(time.Second)
	store.value.State, store.value.Revision = StateOutcomeUnknown, 3
	store.value.DatabaseNow = now
	return Transition{
		State: store.value.State, Revision: store.value.Revision, DatabaseNow: now,
	}, nil
}

func (store *credentialStore) DeleteSecret(
	_ context.Context,
	_ StateCommand,
) (Transition, error) {
	now := store.value.DatabaseNow.Add(time.Second)
	store.value.State, store.value.Revision = StateSecretDeleted, 4
	store.value.DatabaseNow, store.value.SecretDeletedAt = now, &now
	return Transition{
		State: store.value.State, Revision: store.value.Revision, DatabaseNow: now,
	}, nil
}

func (store *credentialStore) Complete(
	_ context.Context,
	_ StateCommand,
) (Transition, error) {
	now := store.value.DatabaseNow.Add(time.Second)
	store.value.State, store.value.Revision = StateCompleted, 5
	store.value.DatabaseNow, store.value.CompletedAt = now, &now
	return Transition{
		State: store.value.State, Revision: store.value.Revision, DatabaseNow: now,
	}, nil
}

var _ Store = (*credentialStore)(nil)
var _ = credentialdeliveryv1.CauseOutcomeUnknown
