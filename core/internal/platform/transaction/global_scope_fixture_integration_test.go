//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

const globalActorID = "16161616-1616-4616-8616-161616161616"

type globalScopeProbe struct {
	database postgres.BoundExecutor
}

func globalScopeProbeBinder(t *testing.T) postgres.GlobalOwnerBinder[globalScopeProbe] {
	t.Helper()
	binder, err := postgres.NewGlobalOwnerBinder(
		"billing", func(database postgres.BoundExecutor) globalScopeProbe {
			return globalScopeProbe{database: database}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return binder
}

func globalCommandRequestFor(
	t *testing.T,
	groupGeneration int64,
	billingGeneration int64,
) GlobalScopeRequest {
	t.Helper()
	capability, err := NewGlobalCommandCapabilityIssuer().Issue(
		uuid.MustParse(globalActorID), "billing.catalog.publish",
		"platform.billing.catalog.manage", 11, 13,
		groupGeneration, billingGeneration,
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := NewGlobalCommandScopeRequest(capability)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func globalPublicRequestFor(
	t *testing.T,
	groupGeneration int64,
	billingGeneration int64,
) GlobalScopeRequest {
	t.Helper()
	capability, err := NewPublicGlobalSnapshotCapabilityIssuer().Issue(
		PublicBillingCatalogList, groupGeneration, billingGeneration,
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := NewPublicGlobalSnapshotScopeRequest(capability)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func (probe globalScopeProbe) searchPath(ctx context.Context) ([]string, error) {
	var value []string
	err := probe.database.QueryRow(
		ctx, "SELECT pg_catalog.current_schemas(false)",
	).Scan(&value)
	return value, err
}

func (probe globalScopeProbe) insert(
	ctx context.Context,
	id string,
	value string,
) error {
	_, err := probe.database.Execute(
		ctx, "INSERT INTO platform.billing_global_probe(id,value) VALUES($1,$2)",
		id, value,
	)
	return err
}

func (probe globalScopeProbe) value(ctx context.Context, id string) (string, error) {
	var value string
	err := probe.database.QueryRow(
		ctx, "SELECT value FROM platform.billing_global_probe WHERE id=$1", id,
	).Scan(&value)
	return value, err
}
