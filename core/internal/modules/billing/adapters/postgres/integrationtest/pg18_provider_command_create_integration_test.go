//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestProviderCommandReplayConflictIsolationAndRollbackOnPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := prepareProviderPG18(t, ctx)
	fixture := prepareProviderPGFixture(t, ctx, database)

	transaction, err := fixture.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, replay, err := submitProviderPGCommand(
		ctx, transaction, fixture, fixture.commandID, fixture.finalizerID,
		fixture.tenantID, fixture.authorityID, fixture.subscriptionID,
		"rollback-effect", fixture.digest,
	); err != nil || replay {
		t.Fatalf("rollback command replay=%t err=%v", replay, err)
	}
	insertProviderPGActiveJob(t, ctx, transaction, fixture)
	insertProviderPGFinalizer(t, ctx, transaction, fixture, fixture.finalizerID)
	if err = transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertProviderPGRollbackEmpty(t, ctx, fixture)

	createProviderPGCommand(t, ctx, &fixture)
	commandID, finalizerID, replay, err := submitProviderPGCommand(
		ctx, fixture.connection, fixture, billingPGV7(t), billingPGV7(t),
		fixture.tenantID, fixture.authorityID, fixture.subscriptionID,
		"provider-effect", fixture.digest,
	)
	if err != nil || !replay || commandID != fixture.commandID ||
		finalizerID != fixture.finalizerID {
		t.Fatalf("command replay ids=%s/%s replay=%t err=%v",
			commandID, finalizerID, replay, err)
	}
	_, _, _, err = submitProviderPGCommand(
		ctx, fixture.connection, fixture, billingPGV7(t), billingPGV7(t),
		fixture.tenantID, fixture.authorityID, fixture.subscriptionID,
		"provider-effect", sha256Bytes("conflict"),
	)
	assertProviderPGCode(t, err, "P5730")

	secondTenant, secondAuthority, secondSubscription :=
		seedProviderPGSecondTenant(t, ctx, fixture)
	_, _, _, err = submitProviderPGCommand(
		ctx, fixture.connection, fixture, billingPGV7(t), billingPGV7(t),
		secondTenant, secondAuthority, secondSubscription,
		"provider-effect", fixture.digest,
	)
	assertProviderPGCode(t, err, "P5730")
	secondCommand, _, secondReplay, err := submitProviderPGCommand(
		ctx, fixture.connection, fixture, billingPGV7(t), billingPGV7(t),
		secondTenant, secondAuthority, secondSubscription,
		"second-tenant-effect", fixture.digest,
	)
	if err != nil || secondReplay || secondCommand == fixture.commandID {
		t.Fatalf("second tenant command=%s replay=%t err=%v",
			secondCommand, secondReplay, err)
	}
}

func submitProviderPGCommand(
	ctx context.Context,
	runner providerPGRunner,
	f providerPGFixture,
	commandID, finalizerID, tenantID, authorityID, subscriptionID uuid.UUID,
	effectKey string,
	digest []byte,
) (uuid.UUID, uuid.UUID, bool, error) {
	var storedCommand, storedFinalizer uuid.UUID
	var replay bool
	var mappingID any
	if tenantID == f.tenantID {
		mappingID = f.mappingID
	}
	err := runner.QueryRowContext(ctx, `
SELECT command_id,finalizer_job_id,replay
FROM platform.billing_create_provider_command_v1(
 $1,$2,$3,$4,$5,$6,$7,'change',$8,$9,$10
)`, authorityID, tenantID, commandID, finalizerID, f.actorID,
		mappingID, subscriptionID, effectKey, digest, f.now,
	).Scan(&storedCommand, &storedFinalizer, &replay)
	return storedCommand, storedFinalizer, replay, err
}

func assertProviderPGRollbackEmpty(
	t *testing.T, ctx context.Context, f providerPGFixture,
) {
	t.Helper()
	var commands, jobs int
	if err := f.connection.QueryRowContext(ctx, `
SELECT (SELECT count(*) FROM platform.billing_provider_commands WHERE id=$1),
       (SELECT count(*) FROM platform.jobs WHERE id IN ($1,$2))
`, f.commandID, f.finalizerID).Scan(&commands, &jobs); err != nil {
		t.Fatal(err)
	}
	if commands != 0 || jobs != 0 {
		t.Fatalf("rollback retained commands=%d jobs=%d", commands, jobs)
	}
}

func assertProviderPGCode(t *testing.T, err error, code string) {
	t.Helper()
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != code {
		t.Fatalf("PostgreSQL error=%v want=%s", err, code)
	}
}
