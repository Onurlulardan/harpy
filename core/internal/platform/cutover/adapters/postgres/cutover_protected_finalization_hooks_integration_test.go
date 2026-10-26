//go:build integration

package postgres

import (
	"context"
	"testing"
)

type protectedFinalizationHooks struct {
	beforeAuthorization func(builtCutoverLeaseArtifacts, finalizationEvidence)
	afterAuthorization  func(builtCutoverLeaseArtifacts, finalizationEvidence)
	afterVerification   func(builtCutoverLeaseArtifacts, finalizationEvidence)
}

func assertMissingProtectedAuthorization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	connection := connectCutoverRole(
		t, ctx, database.dsn, "harpy_cutover_finalizer",
	)
	defer connection.Close(ctx)
	if _, err := callFinalization(ctx, connection, evidence); err == nil {
		t.Fatal("protected finalization without external authorization succeeded")
	}
}
