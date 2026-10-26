package providerapp

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

type concurrentWebhookIDs struct{}

func (concurrentWebhookIDs) New() uuid.UUID { return uuid.New() }

type concurrentProtector struct{}

func (concurrentProtector) Protect(
	_ context.Context,
	_ []byte,
	body []byte,
) (ProtectedBody, error) {
	return (&webhookProtector{}).Protect(context.Background(), []byte("aad"), body)
}

type concurrentReceipt struct {
	id     uuid.UUID
	digest [32]byte
}

type concurrentIngressStore struct {
	mutex    sync.Mutex
	receipts map[string]concurrentReceipt
}

func (store *concurrentIngressStore) AcceptWebhook(
	_ context.Context,
	acceptance WebhookAcceptance,
) (WebhookAcceptDecision, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	key := string(acceptance.Provider()) + "\x00" + acceptance.Verified().EventID()
	existing, found := store.receipts[key]
	if found {
		if existing.digest != acceptance.Body().BodySHA256() {
			return WebhookAcceptDecision{}, ErrWebhookConflict
		}
		return WebhookAcceptDecision{
			ReceiptID: existing.id, TenantID: webhookTenantID, Created: false,
		}, nil
	}
	store.receipts[key] = concurrentReceipt{
		id: acceptance.ReceiptID(), digest: acceptance.Body().BodySHA256(),
	}
	return WebhookAcceptDecision{
		ReceiptID: acceptance.ReceiptID(), TenantID: webhookTenantID, Created: true,
	}, nil
}

func TestReceiveWebhookConcurrentReplayHasOneDurableReceipt(t *testing.T) {
	verified := verifiedWebhookFixture(t)
	registry, key, _ := webhookRegistryFixture(
		t, verified, normalizedWebhookFixture(t, verified),
	)
	store := &concurrentIngressStore{receipts: make(map[string]concurrentReceipt)}
	ingress, _ := NewIngress(
		registry, webhookClock{webhookNow}, concurrentWebhookIDs{},
		concurrentProtector{}, store,
	)
	headers, _ := NewHeaders(map[string][]string{"Testpay-Signature": {"signed"}})
	command, _ := NewWebhookCommand(key, headers, []byte(`{"id":"evt_100"}`))
	const workers = 32
	results := make(chan ReceiveResult, workers)
	errorsSeen := make(chan error, workers)
	var created atomic.Int32
	var wait sync.WaitGroup
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := ingress.ReceiveWebhook(context.Background(), command)
			if result.Created {
				created.Add(1)
			}
			results <- result
			errorsSeen <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errorsSeen)
	var receiptID uuid.UUID
	for err := range errorsSeen {
		if err != nil {
			t.Fatalf("concurrent receive error=%v", err)
		}
	}
	for result := range results {
		if receiptID == uuid.Nil {
			receiptID = result.ReceiptID
		}
		if result.ReceiptID != receiptID {
			t.Fatalf("receipt drift: %s != %s", result.ReceiptID, receiptID)
		}
	}
	if created.Load() != 1 || len(store.receipts) != 1 {
		t.Fatalf("created=%d receipts=%d", created.Load(), len(store.receipts))
	}
	different, _ := NewWebhookCommand(key, headers, []byte(`{"id":"evt_100","x":1}`))
	if _, err := ingress.ReceiveWebhook(context.Background(), different); !errors.Is(
		err, ErrWebhookConflict,
	) {
		t.Fatalf("different body reuse error=%v", err)
	}
}
