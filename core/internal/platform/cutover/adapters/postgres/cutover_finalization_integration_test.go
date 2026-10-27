//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func verifyFinalizationContract(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
) {
	t.Helper()
	verifyProtectedFinalizationWorkflow(t, ctx, database)
}

type finalizationOutcome struct {
	state          string
	evidenceSHA256 string
	err            error
}

func raceFinalization(
	ctx context.Context,
	first *pgx.Conn,
	second *pgx.Conn,
	evidence finalizationEvidence,
) [2]finalizationOutcome {
	start := make(chan struct{})
	outcomes := [2]finalizationOutcome{}
	connections := []*pgx.Conn{first, second}
	var group sync.WaitGroup
	for index := range connections {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			outcome, err := callFinalization(ctx, connections[index], evidence)
			outcome.err = err
			outcomes[index] = outcome
		}(index)
	}
	close(start)
	group.Wait()
	return outcomes
}

func callFinalization(
	ctx context.Context,
	connection finalizationQueryRow,
	evidence finalizationEvidence,
) (finalizationOutcome, error) {
	var outcome finalizationOutcome
	err := connection.QueryRow(ctx, finalizeOperationSQL,
		evidence.operationID, evidence.runID, evidence.leaseID,
		evidence.leaseSHA256, evidence.leaseDocument,
		evidence.adoptionSHA256, evidence.adoptionDocument,
		evidence.sealSHA256, evidence.sealDocument,
		evidence.cleanupSHA256, evidence.cleanupDocument,
		evidence.authoritySHA256, evidence.authorityDocument,
		evidence.terminalState, evidence.terminalSHA256, evidence.terminalDocument,
	).Scan(&outcome.state, &outcome.evidenceSHA256)
	outcome.err = err
	return outcome, err
}

type finalizationQueryRow interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func replaceFinalizerLeaseField(
	t *testing.T,
	document json.RawMessage,
	field string,
	value string,
) json.RawMessage {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatal(err)
	}
	decoded[field] = value
	return mustIntegrationJSON(t, decoded)
}

func assertFinalizationError(t *testing.T, err error, code string, message string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != code ||
		!strings.Contains(postgresError.Message, message) {
		t.Fatalf("finalization error = %v", err)
	}
}
