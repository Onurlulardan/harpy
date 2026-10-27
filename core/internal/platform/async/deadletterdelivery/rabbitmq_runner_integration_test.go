//go:build integration

package deadletterdelivery

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
)

type managedIngestStore struct {
	mutex     *sync.Mutex
	calls     int
	failFirst bool
	persisted chan deadletter.EventRoot
}

func newManagedIngestStore() *managedIngestStore {
	return &managedIngestStore{
		mutex: &sync.Mutex{}, failFirst: true,
		persisted: make(chan deadletter.EventRoot, 4),
	}
}

func (store *managedIngestStore) IngestEvent(
	_ context.Context,
	root deadletter.EventRoot,
) (deadletter.IngestResult, error) {
	store.mutex.Lock()
	store.calls++
	shouldFail := store.failFirst
	store.failFirst = false
	store.mutex.Unlock()
	if shouldFail {
		return deadletter.IngestResult{}, errors.New("database unavailable")
	}
	store.persisted <- root
	return deadletter.IngestResult{
		RootID: root.RootID(), Created: true, RecordedAt: time.Now().UTC(),
	}, nil
}

func (store *managedIngestStore) Calls() int {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.calls
}

func TestRabbitMQ432ManagedDeadLetterStoreFailureAndSessionRestart(t *testing.T) {
	fixture := startManagedRabbit(t)
	catalog := managedConsumerCatalog(t)
	store := newManagedIngestStore()
	connector, err := NewRabbitConnector(fixture.settings, catalog, store)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := New(connector, Settings{
		ReconnectMinimum: fixture.settings.ReconnectMinimum(),
		ReconnectMaximum: fixture.settings.ReconnectMaximum(),
	}, openDeadLetterAdmission())
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	firstID := "11111111-1111-4111-8111-111111111111"
	publishManagedDeadLetter(t, fixture.settings, catalog, managedEnvelopeDocument(t, firstID))
	first := waitManagedRoot(t, store.persisted)
	if first.Envelope().EventID != firstID || store.Calls() != 2 {
		t.Fatalf("first delivery event=%s store calls=%d", first.Envelope().EventID, store.Calls())
	}
	time.Sleep(100 * time.Millisecond)
	restartManagedSession(t, runner)
	secondID := "66666666-6666-4666-8666-666666666666"
	publishManagedDeadLetter(t, fixture.settings, catalog, managedEnvelopeDocument(t, secondID))
	second := waitManagedRoot(t, store.persisted)
	if second.Envelope().EventID != secondID || store.Calls() != 3 {
		t.Fatalf("restart delivery event=%s store calls=%d", second.Envelope().EventID, store.Calls())
	}
	runner.BeginDrain()
	waitContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := runner.Wait(waitContext); err != nil {
		t.Fatal(err)
	}
	if failure, open := <-runner.Failures(); open || failure != nil {
		t.Fatalf("managed runner failure=%v open=%t", failure, open)
	}
}

func restartManagedSession(t *testing.T, runner *Runner) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		runner.lifecycle.mutex.Lock()
		active := runner.lifecycle.active
		runner.lifecycle.mutex.Unlock()
		if active != nil {
			if err := active.Close(); err != nil {
				t.Fatal(err)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("managed RabbitMQ session did not become active")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func waitManagedRoot(
	t *testing.T,
	results <-chan deadletter.EventRoot,
) deadletter.EventRoot {
	t.Helper()
	select {
	case root := <-results:
		return root
	case <-time.After(30 * time.Second):
		t.Fatal("managed dead-letter was not durably ingested")
		return deadletter.EventRoot{}
	}
}
