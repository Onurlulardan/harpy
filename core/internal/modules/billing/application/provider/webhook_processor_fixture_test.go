package providerapp

import (
	"context"

	"github.com/google/uuid"
)

type bodyFixtureOpener struct {
	body []byte
	err  error
}

func (opener *bodyFixtureOpener) OpenForReconciliation(
	_ context.Context,
	_ []byte,
	_ ProtectedBody,
) ([]byte, error) {
	return append([]byte(nil), opener.body...), opener.err
}

type processorFixtureStore struct {
	stored      StoredWebhook
	application WebhookApplication
	decision    WebhookProcessDecision
	err         error
	applyCalls  int
}

func (store *processorFixtureStore) LoadWebhook(
	_ context.Context,
	_ uuid.UUID,
) (StoredWebhook, error) {
	return store.stored, store.err
}

func (store *processorFixtureStore) ApplyWebhook(
	_ context.Context,
	application WebhookApplication,
) (WebhookProcessDecision, error) {
	store.applyCalls++
	store.application = application
	return store.decision, store.err
}
