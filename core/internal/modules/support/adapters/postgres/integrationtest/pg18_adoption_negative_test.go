//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

type stopBeforeReconcile struct{ application.Target }

func (stopBeforeReconcile) AdoptionReconcile(context.Context,
	application.LeaseCommand,
) (application.Run, error) {
	return application.Run{}, application.ErrUnavailable
}

func exerciseTargetChecksumMismatch(t *testing.T, ctx context.Context, database *sql.DB,
	store application.Target, clock *adoptionClock, locator application.VerifiedLocator,
) {
	t.Helper()
	row := contactRow("06000000-0000-4000-8000-000000000001", "TargetDrift", clock.value)
	source := contactSourceFixture(t, locator, []application.ContactSourceRow{row}, "target-drift")
	worker, _ := application.NewWorker(source, stopBeforeReconcile{Target: store}, clock)
	run, _, err := worker.Start(ctx, application.StartRequest{
		OperationID: uuid.New(), RunID: uuid.Must(uuid.NewV7()), Locator: locator})
	if err != nil {
		t.Fatal(err)
	}
	leaseOwner := uuid.New()
	_, err = worker.Execute(ctx, application.ProcessRequest{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: leaseOwner, ExpectedRevision: run.Revision})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("reconcile stop error=%v", err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE platform.support_contact_messages AS message
        SET name='Changed after copy'
        FROM platform.support_adoption_mappings AS mapping
        WHERE mapping.run_id=$1 AND mapping.target_id=message.id`, run.RunID); err != nil {
		t.Fatal(err)
	}
	claimed, err := store.AdoptionClaim(ctx, application.ClaimCommand{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: leaseOwner, ExpectedRevision: 3, ObservedAt: clock.value})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.AdoptionReconcile(ctx, application.LeaseCommand{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: leaseOwner, LeaseGeneration: claimed.LeaseGeneration,
		ObservedAt: clock.value})
	if !errors.Is(err, application.ErrSourceDrift) {
		t.Fatalf("target checksum error=%v", err)
	}
}

func exerciseLocatorMismatch(t *testing.T, ctx context.Context, store application.Target,
	clock *adoptionClock, locator application.VerifiedLocator,
) {
	t.Helper()
	proof := sha256.Sum256([]byte("wrong-catalog-locator"))
	wrong, err := application.NewVerifiedLocator(application.ContactGlobal, uuid.Nil,
		locator.Schema(), locator.Relation(), locator.RelationOID()+1, "", "", proof)
	if err != nil {
		t.Fatal(err)
	}
	source := contactSourceFixture(t, wrong, nil, "wrong-locator")
	worker, _ := application.NewWorker(source, store, clock)
	_, _, err = worker.Start(ctx, application.StartRequest{
		OperationID: uuid.New(), RunID: uuid.Must(uuid.NewV7()), Locator: wrong})
	if !errors.Is(err, application.ErrSourceDrift) {
		t.Fatalf("catalog locator error=%v", err)
	}
}
