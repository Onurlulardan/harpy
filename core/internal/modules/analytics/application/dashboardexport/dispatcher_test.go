package dashboardexport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRenderDispatcherReservesSpoolBeforeExactClaim(t *testing.T) {
	t.Parallel()
	dispatcher, claims, spools, events := renderDispatcherFixture(nil, nil)
	group := newRenderDispatchGroup(func() {})
	worked, err := dispatcher.claimAndLaunch(context.Background(),
		context.Background(), group)
	group.wait.Wait()
	if err != nil || !worked || events.joined() != "peek,reserve,claim,execute" ||
		claims.claims != 1 || spools.reserved == nil || !spools.reserved.discarded {
		t.Fatalf("dispatch = %v/%v events=%s claims=%d spool=%#v",
			worked, err, events.joined(), claims.claims, spools.reserved)
	}
}

func TestRenderDispatcherSpoolCapacityDoesNotClaimJob(t *testing.T) {
	t.Parallel()
	dispatcher, claims, _, events := renderDispatcherFixture(ErrSpoolCapacity, nil)
	worked, err := dispatcher.claimAndLaunch(context.Background(),
		context.Background(), newRenderDispatchGroup(func() {}))
	if err != nil || worked || claims.claims != 0 || events.joined() != "peek,reserve" {
		t.Fatalf("dispatch = %v/%v events=%s claims=%d",
			worked, err, events.joined(), claims.claims)
	}
}

func TestRenderDispatcherKnownClaimWaitDiscardsReservation(t *testing.T) {
	t.Parallel()
	dispatcher, claims, spools, _ := renderDispatcherFixture(nil, ErrDispatchQuotaWait)
	worked, err := dispatcher.claimAndLaunch(context.Background(),
		context.Background(), newRenderDispatchGroup(func() {}))
	if err != nil || worked || claims.claims != 1 ||
		spools.reserved == nil || !spools.reserved.discarded {
		t.Fatalf("dispatch = %v/%v claims=%d spool=%#v",
			worked, err, claims.claims, spools.reserved)
	}
}

func TestRenderDispatcherAmbiguousClaimRetainsReservation(t *testing.T) {
	t.Parallel()
	dispatcher, _, spools, _ := renderDispatcherFixture(nil, errors.New("commit unknown"))
	worked, err := dispatcher.claimAndLaunch(context.Background(),
		context.Background(), newRenderDispatchGroup(func() {}))
	if err == nil || worked || spools.reserved == nil || spools.reserved.discarded {
		t.Fatalf("dispatch = %v/%v spool=%#v", worked, err, spools.reserved)
	}
}

func TestRenderDispatcherUnknownOutcomeRetainsSpool(t *testing.T) {
	t.Parallel()
	dispatcher, claims, spools, _ := renderDispatcherFixture(nil, nil)
	dispatcher.executor = unknownDispatcherExecutor{}
	reservation := SpoolReservationIdentity{ExportID: claims.candidate.ExportID,
		AttemptGeneration: 1,
		LeaseOwnerID:      uuid.MustParse("019b1000-0000-7000-8000-000000000034")}
	spool, _ := spools.Reserve(reservation)
	attempt, _ := claims.ClaimRenderCandidate(context.Background(),
		claims.candidate, reservation)
	if err := dispatcher.executeWithHeartbeats(context.Background(),
		renderDispatch{attempt: attempt, spool: spool}, make(chan time.Time)); err != nil ||
		spools.reserved.discarded {
		t.Fatalf("unknown outcome = %v discarded=%v", err, spools.reserved.discarded)
	}
}

func TestRenderDispatcherLeaseLossCancelsExecutionAndRetainsSpool(t *testing.T) {
	t.Parallel()
	dispatcher, claims, spools, _ := renderDispatcherFixture(nil, nil)
	dispatcher.executor = blockingDispatcherExecutor{}
	claims.checkpointErr = ErrDispatchLeaseStale
	reservation := SpoolReservationIdentity{ExportID: claims.candidate.ExportID,
		AttemptGeneration: 1,
		LeaseOwnerID:      uuid.MustParse("019b1000-0000-7000-8000-000000000034")}
	spool, _ := spools.Reserve(reservation)
	attempt, _ := claims.ClaimRenderCandidate(context.Background(),
		claims.candidate, reservation)
	heartbeats := make(chan time.Time, 1)
	heartbeats <- dispatcherTime.Add(20 * time.Second)
	if err := dispatcher.executeWithHeartbeats(context.Background(),
		renderDispatch{attempt: attempt, spool: spool}, heartbeats); err != nil ||
		claims.checkpoints != 1 || spools.reserved.discarded {
		t.Fatalf("lease loss = %v checkpoints=%d discarded=%v", err,
			claims.checkpoints, spools.reserved.discarded)
	}
}
