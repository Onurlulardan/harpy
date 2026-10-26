//go:build integration

package api

import (
	"context"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres/integrationfixture"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func assertCatalogDemoAnalyticsFormsFlow(
	t *testing.T,
	ctx context.Context,
	root string,
	environment *integrationfixture.CatalogDemoEnvironment,
	coordinator *platformtx.Coordinator,
	formsOwner any,
	analyticsOwner any,
) {
	t.Helper()
	formsRequest := catalogDemoInventoryFormsRequest(t, root, formsOwner)
	analyticsRequest := catalogDemoInventoryAnalyticsRequest(
		t, root, analyticsOwner,
	)
	analyticsPending := catalogDemoProvision(
		t, ctx, coordinator, analyticsOwner, analyticsRequest,
	)
	assertCatalogDemoAnalyticsPending(t, analyticsPending)
	environment.AssertCatalogDemoAnalyticsReferences(
		t, ctx, catalogDemoAnalyticsComponentID,
		catalogDemoInventoryBinID, 0,
	)

	formsPending := catalogDemoProvision(
		t, ctx, coordinator, formsOwner, formsRequest,
	)
	assertCatalogDemoPending(t, "Inventory Forms", formsPending)
	environment.CompleteCatalogDemoAnalyticsFormsApply(
		t, ctx, catalogDemoInventoryBinID,
	)
	formsReady := catalogDemoProvision(
		t, ctx, coordinator, formsOwner, formsRequest,
	)
	assertCatalogDemoReadyReplay(
		t, "Inventory Forms", formsPending, formsReady,
	)
	catalogDemoActivateForms(
		t, ctx, coordinator, formsOwner, formsRequest,
	)

	analyticsReady := catalogDemoProvision(
		t, ctx, coordinator, analyticsOwner, analyticsRequest,
	)
	assertCatalogDemoReadyReplay(
		t, "Analytics", analyticsPending, analyticsReady,
	)
	evidence := environment.AssertCatalogDemoAnalyticsReferences(
		t, ctx, catalogDemoAnalyticsComponentID,
		catalogDemoInventoryBinID, 1,
	)
	analyticsReplay := catalogDemoProvision(
		t, ctx, coordinator, analyticsOwner, analyticsRequest,
	)
	assertCatalogDemoExactReadyReplay(
		t, analyticsReady, analyticsReplay,
	)
	replayEvidence := environment.AssertCatalogDemoAnalyticsReferences(
		t, ctx, catalogDemoAnalyticsComponentID,
		catalogDemoInventoryBinID, 1,
	)
	if !reflect.DeepEqual(evidence, replayEvidence) {
		t.Fatalf(
			"Analytics Forms replay duplicated evidence: before=%+v after=%+v",
			evidence, replayEvidence,
		)
	}
}

func assertCatalogDemoAnalyticsPending(
	t *testing.T, receipt catalogDemoProvisionReceipt,
) {
	t.Helper()
	if !catalogDemoUUIDv7(receipt.ID) || receipt.Ready ||
		receipt.State != "STAGED" || receipt.RetryAt.IsZero() ||
		receipt.CompletedAt.IsZero() ||
		!receipt.RetryAt.After(receipt.CompletedAt) {
		t.Fatalf("Analytics pending receipt differs: %+v", receipt)
	}
}
