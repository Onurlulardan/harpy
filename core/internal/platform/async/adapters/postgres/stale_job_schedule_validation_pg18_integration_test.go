//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertStaleDispositionRejectsNullSchedules(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	tenantID := addStaleJobTenant(t, ctx, fixture)
	jobID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectDatabaseTransactional, 8,
	)
	setPublisherDomainGeneration(t, ctx, fixture.publisherPGFixture, tenantID, 2)
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 1 {
		t.Fatalf("park null-schedule candidate = %d, %v", parked, err)
	}
	authority, err := fixture.store.ResolveStaleJobAuthority(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	command := staleDispositionCommand(
		t, jobID, jobsv1.StaleJobAdopt, "validate schedule evidence",
	)
	for _, test := range []struct {
		name      string
		argument  int
		wantState string
	}{
		{name: "source_run_at", argument: 20, wantState: "P5687"},
		{name: "result_run_at", argument: 21, wantState: "P5687"},
		{name: "occurred_at", argument: 22, wantState: "P5684"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, callErr := withJobAuthorityTransaction(
				ctx, fixture.store, authority, func(transaction pgx.Tx) (bool, error) {
					source, err := inspectStaleJobDisposition(
						ctx, transaction, authority, command.jobID,
					)
					if err != nil {
						return false, err
					}
					prepared, err := source.prepare(authority, command)
					if err != nil {
						return false, err
					}
					arguments := staleDispositionArguments(authority, command, prepared)
					arguments[test.argument] = nil
					var completed any
					err = transaction.QueryRow(ctx, disposeStaleJobSQL, arguments...).Scan(
						&completed, new(uuid.UUID), new(uuid.UUID), new(uuid.UUID), new(bool),
					)
					return false, err
				},
			)
			var postgresError *pgconn.PgError
			if !errors.As(callErr, &postgresError) || postgresError.Code != test.wantState {
				t.Fatalf("null %s error = %v", test.name, callErr)
			}
		})
	}
}

func staleDispositionArguments(
	authority JobAuthority,
	command StaleJobDispositionCommand,
	prepared preparedStaleJobDisposition,
) []any {
	return []any{
		command.jobID, string(command.decision), prepared.rootID, prepared.actionID,
		optionalDispositionUUID(prepared.resultJobID),
		optionalDispositionUUID(prepared.resultRootID),
		authority.scopeKind, authority.scopeID, authority.groupID,
		authority.groupGeneration, authority.domain, authority.generation,
		optionalDispositionText(prepared.resultPayloadReference),
		optionalDispositionText(prepared.resultDedupeKey), prepared.canonical,
		optionalDispositionText(prepared.digest), command.reason, command.ticket,
		command.actorID, command.causationID, prepared.sourceRunAt,
		optionalDispositionText(prepared.resultRunAt), prepared.occurredAt,
	}
}
