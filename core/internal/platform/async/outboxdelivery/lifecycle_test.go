package outboxdelivery

import (
	"context"
	"errors"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

type unavailableConnector struct{}

func (unavailableConnector) Open(context.Context) (Broker, error) {
	return nil, errors.New("unavailable")
}

type unreachableStore struct{}

func (unreachableStore) AssertOutboxCatalog(
	context.Context,
	eventsv1.ProducerCatalog,
) error {
	return errors.New("unexpected store access")
}

func (unreachableStore) BeginOutboxAuthorityPage(int) (asyncpostgres.OutboxAuthorityPageToken, error) {
	return asyncpostgres.OutboxAuthorityPageToken{}, errors.New("unexpected store access")
}
func (unreachableStore) ListOutboxAuthorities(
	context.Context, asyncpostgres.OutboxAuthorityPageToken,
) (asyncpostgres.OutboxAuthorityPage, error) {
	return asyncpostgres.OutboxAuthorityPage{}, errors.New("unexpected store access")
}
func (unreachableStore) ClaimOutbox(
	context.Context, asyncpostgres.OutboxAuthority, uuid.UUID, int, eventsv1.ProducerCatalog,
) ([]asyncpostgres.PublishClaim, error) {
	return nil, errors.New("unexpected store access")
}
func (unreachableStore) HeartbeatPublish(
	context.Context, asyncpostgres.PublishClaim,
) (asyncpostgres.PublishLease, error) {
	return asyncpostgres.PublishLease{}, errors.New("unexpected store access")
}
func (unreachableStore) ConfirmPublish(
	context.Context, asyncpostgres.PublishClaim, string,
) (time.Time, error) {
	return time.Time{}, errors.New("unexpected store access")
}
func (unreachableStore) ReleasePublish(
	context.Context, asyncpostgres.PublishClaim, time.Duration, string, string,
) (time.Time, error) {
	return time.Time{}, errors.New("unexpected store access")
}
func (unreachableStore) RecoverExpiredOutbox(
	context.Context, asyncpostgres.OutboxAuthority, int,
) (int64, error) {
	return 0, errors.New("unexpected store access")
}
func (unreachableStore) ParkStaleOutbox(
	context.Context, asyncpostgres.OutboxAuthority, int,
) (int64, error) {
	return 0, errors.New("unexpected store access")
}

func TestRunnerReconnectLoopDrainsWithoutLeaking(t *testing.T) {
	catalog, _ := eventsv1.NewProducerCatalog(nil)
	runner, err := newWithJitter(
		unreachableStore{}, catalog, unavailableConnector{}, Settings{
			ConfirmTimeout: time.Second, ReconnectMinimum: time.Millisecond,
			ReconnectMaximum: 4 * time.Millisecond,
		}, openOutboxAdmission(t), fixedJitter(0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); !errors.Is(err, ErrAlreadyStart) {
		t.Fatalf("second start error = %v", err)
	}
	runner.BeginDrain()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if _, open := <-runner.Failures(); open {
		t.Fatal("failure channel remained open after drain")
	}
}

func openOutboxAdmission(t *testing.T) *admission.Gate {
	t.Helper()
	gate := admission.NewGate(nil)
	if !gate.Open() {
		t.Fatal("outbox admission did not open")
	}
	return gate
}
