//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type scopeProbe struct {
	database postgres.BoundExecutor
}

func scopeProbeBinder(t *testing.T) postgres.OwnerBinder[scopeProbe] {
	t.Helper()
	binder, err := postgres.NewOwnerBinder("iam", func(database postgres.BoundExecutor) scopeProbe {
		return scopeProbe{database: database}
	})
	if err != nil {
		t.Fatal(err)
	}
	return binder
}

func scopeRequestFor(t *testing.T, tenant string, generation int64) ScopeRequest {
	t.Helper()
	actor, err := NewActorIssuer().Issue(
		uuid.MustParse("11111111-1111-4111-8111-111111111111"), uuid.MustParse(tenant),
	)
	if err != nil {
		t.Fatal(err)
	}
	operation, err := NewAuthorityIssuer().Issue(
		"iam.probe.create", "iam_system_resources", generation,
		map[string]int64{"iam": generation, "system_resource_layout": generation},
	)
	if err != nil {
		t.Fatal(err)
	}
	return ScopeRequest{Actor: actor, Operation: operation, Mode: Command}
}

func scopeSnapshotRequestFor(
	t *testing.T, tenant string, generations map[string]int64,
) ScopeRequest {
	t.Helper()
	actor, err := NewActorIssuer().Issue(
		uuid.MustParse("11111111-1111-4111-8111-111111111111"), uuid.MustParse(tenant),
	)
	if err != nil {
		t.Fatal(err)
	}
	operation, err := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", generations,
	)
	if err != nil {
		t.Fatal(err)
	}
	return ScopeRequest{Actor: actor, Operation: operation, Mode: SnapshotQuery}
}

func (probe scopeProbe) count(ctx context.Context) (int, error) {
	var count int
	err := probe.database.QueryRow(ctx, "SELECT count(*) FROM iam_probe").Scan(&count)
	return count, err
}

func (probe scopeProbe) backendPID(ctx context.Context) (int32, error) {
	var pid int32
	err := probe.database.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid)
	return pid, err
}
