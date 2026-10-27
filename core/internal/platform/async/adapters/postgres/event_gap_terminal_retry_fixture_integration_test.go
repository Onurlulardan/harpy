//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type terminalGapRetryCase struct {
	fixture       *eventGapPGFixture
	rootID        uuid.UUID
	jobID         uuid.UUID
	terminalRoot  uuid.UUID
	sourceEventID uuid.UUID
}

func newTerminalGapRetryCase(
	t *testing.T,
	ctx context.Context,
	base *eventGapPGFixture,
) terminalGapRetryCase {
	t.Helper()
	fixture := *base
	fixture.aggregate = uuid.New()
	fixture.envelopes = make(map[int64]eventsv1.Envelope)
	now := time.Now().UTC()
	fixture.seedDeliveredAt(t, ctx, 1, now.Add(-33*24*time.Hour))
	fixture.seedDeliveredAt(t, ctx, 2, now.Add(-32*24*time.Hour))
	if fixture.deliver(t, ctx, 2) != brokerrabbit.HandleCommitted {
		t.Fatal("event gap was not parked")
	}
	rootID, jobID := fixture.gapIDs(t, ctx, 2)
	if fixture.deliver(t, ctx, 1) != brokerrabbit.HandleCommitted {
		t.Fatal("event gap predecessor was not committed")
	}
	claim := fixture.claimGap(t, ctx, jobID)
	terminalizeGapClaim(t, ctx, &fixture, claim)
	var terminalRoot uuid.UUID
	if err := fixture.runtime.QueryRow(ctx, `
SELECT id FROM platform.dead_letters
WHERE record_kind='root' AND source_kind='job' AND source_id=$1`,
		jobID,
	).Scan(&terminalRoot); err != nil {
		t.Fatalf("read terminal event-gap job root: %v", err)
	}
	return terminalGapRetryCase{
		fixture: &fixture, rootID: rootID, jobID: jobID,
		terminalRoot:  terminalRoot,
		sourceEventID: uuid.MustParse(fixture.envelopes[2].EventID),
	}
}

func terminalizeGapClaim(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	claim JobClaim,
) {
	t.Helper()
	failure, err := NewJobFailure(
		jobsv1.ErrorPermanent, "EVENT_GAP_TEST_FAILURE",
		strings.Repeat("e", 64), 0,
	)
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(context.Background())
	bound := &repository{
		database:  lifecycleBoundPGTx{transaction: transaction},
		authority: lifecycleAuthorityStamp(claim.Authority()),
	}
	result, err := bound.FailJobBeforeEffect(ctx, claim, failure)
	if err != nil || result.State() != jobsv1.StateDead {
		t.Fatalf("terminalize event-gap claim = %#v: %v", result, err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func retryTerminalGap(
	ctx context.Context,
	value terminalGapRetryCase,
	retryJobID uuid.UUID,
) (deadletter.RetryResult, error) {
	repository := &deadLetterRepository{
		database: value.fixture.runtime, cleanup: 5 * time.Second,
		consumers: value.fixture.consumers,
	}
	operator, err := deadletter.NewOperator(
		deadLetterPermissionFixture{allowed: map[uuid.UUID]struct{}{
			value.fixture.tenantID: {},
		}},
		repository, deadLetterFixedID{value: retryJobID},
	)
	if err != nil {
		return deadletter.RetryResult{}, err
	}
	return operator.Retry(ctx, deadletter.RetryRequest{
		ActorID: uuid.New(), TenantID: value.fixture.tenantID,
		RootID: value.terminalRoot, Reason: "event gap evidence repaired",
		Ticket: "OPS-EVENT-GAP", CausationID: uuid.New(),
	})
}
