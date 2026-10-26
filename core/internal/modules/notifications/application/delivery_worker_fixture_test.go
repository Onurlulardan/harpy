package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type workerStoreFake struct {
	work        ports.DeliveryWork
	finalized   ports.DeliveryFinalization
	finalizeErr error
}

func (store *workerStoreFake) AcquireDelivery(_ context.Context, _, _ domain.ID,
	owner string, now time.Time,
) (ports.DeliveryWork, error) {
	claimed, err := store.work.Delivery.Claim(owner, now)
	store.work.Delivery = claimed
	return store.work, err
}

func (store *workerStoreFake) CommitDelivery(_ context.Context, _ domain.ID,
	final ports.DeliveryFinalization, expected int64,
) error {
	if store.finalizeErr != nil {
		return store.finalizeErr
	}
	if expected != store.work.Delivery.Revision {
		return ErrConflict
	}
	store.finalized = final
	store.work.Delivery = final.Delivery
	return nil
}

type targetFake struct{ value []byte }

func (target targetFake) ResolveDeliveryTarget(context.Context, domain.ID,
	domain.ID, domain.Channel,
) ([]byte, error) {
	return target.value, nil
}

type openerFake struct{ value []byte }

func (opener openerFake) Open(context.Context, ports.EncryptedSecret) ([]byte, error) {
	return opener.value, nil
}

type senderFake struct {
	result     ports.ProviderResult
	target     []byte
	credential []byte
}

func (sender *senderFake) Send(_ context.Context,
	request ports.ProviderRequest,
) ports.ProviderResult {
	sender.target = request.Target
	sender.credential = request.Credential
	return sender.result
}

type registryFake struct{ sender ports.ProviderSender }

func (registry registryFake) Resolve(ports.ProviderKind) (ports.ProviderSender, error) {
	if registry.sender == nil {
		return nil, ports.ErrProviderNotFound
	}
	return registry.sender, nil
}

func workerFixture(t *testing.T, result ports.ProviderResult) (*DeliveryWorker,
	*workerStoreFake, *senderFake, domain.ID, domain.ID,
) {
	t.Helper()
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	ids := &testIDs{}
	tenantID, _ := ids.New(now)
	recipientID, _ := ids.New(now)
	messageID, _ := ids.New(now)
	templateID, _ := ids.New(now)
	providerID, _ := ids.New(now)
	deliveryID, _ := ids.New(now)
	secretID, _ := ids.New(now)
	message, err := domain.NewMessage(domain.Message{ID: messageID,
		RecipientID: recipientID, TemplateVersionID: templateID, Producer: "workflow",
		NotificationType: "workflow.task.due", Category: domain.CategoryWorkflowAction,
		TemplateKind: domain.TemplatePlainText, Locale: "en", Subject: "Task",
		Body: "Review task", Revision: 1,
		InboxRevision: 1, CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	delivery, err := domain.NewDelivery(domain.Delivery{ID: deliveryID,
		MessageID: messageID, ProviderConfigID: providerID, Channel: domain.ChannelEmail,
		State: domain.DeliveryPending, MaximumAttempts: 8, NotBefore: now, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	store := &workerStoreFake{work: ports.DeliveryWork{Delivery: delivery,
		Message: message, Provider: ports.ProviderConfig{ID: providerID,
			VersionID: providerID, SecretID: secretID, Kind: ports.ProviderSMTP,
			Enabled: true, Revision: 1,
			MetadataDigest: sha256.Sum256([]byte("smtp-account"))},
		Secret: ports.EncryptedSecret{ID: secretID}}}
	target, credential := []byte("user@example.test"), []byte("smtp-secret")
	sender := &senderFake{result: result}
	worker, err := NewDeliveryWorker(store, targetFake{target}, openerFake{credential},
		registryFake{sender}, testClock{now}, ids)
	if err != nil {
		t.Fatal(err)
	}
	return worker, store, sender, tenantID, deliveryID
}

func zeroed(value []byte) bool { return bytes.Equal(value, make([]byte, len(value))) }

var errFinalizeFake = errors.New("finalize failed")
