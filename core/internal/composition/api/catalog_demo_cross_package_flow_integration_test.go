//go:build integration

package api

import (
	"context"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres/integrationfixture"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func assertCatalogDemoCrossPackageFlow(
	t *testing.T,
	ctx context.Context,
	root string,
	environment *integrationfixture.CatalogDemoEnvironment,
	coordinator *platformtx.Coordinator,
	formsOwner any,
	catalogOwner any,
	projectRecordID string,
) {
	t.Helper()
	formsRequest := catalogDemoSalesFormsRequest(t, root, formsOwner)
	catalogRequest := catalogDemoSalesCatalogRequest(t, catalogOwner)

	catalogPending := catalogDemoProvision(
		t, ctx, coordinator, catalogOwner, catalogRequest,
	)
	assertCatalogDemoPending(t, "cross-package Catalog", catalogPending)
	formsPending := catalogDemoProvision(
		t, ctx, coordinator, formsOwner, formsRequest,
	)
	assertCatalogDemoPending(t, "cross-package Forms", formsPending)
	environment.CompleteCatalogDemoFormsPublication(
		t, ctx, catalogDemoLostReasonID,
	)
	formsReady := catalogDemoProvision(
		t, ctx, coordinator, formsOwner, formsRequest,
	)
	assertCatalogDemoReadyReplay(
		t, "cross-package Forms", formsPending, formsReady,
	)
	catalogReady := catalogDemoProvision(
		t, ctx, coordinator, catalogOwner, catalogRequest,
	)
	assertCatalogDemoReadyReplay(
		t, "cross-package Catalog", catalogPending, catalogReady,
	)
	evidence := environment.AssertCatalogDemoCrossPackageEvidence(
		t, ctx, catalogDemoSalesPackageID, catalogDemoCrossLegacy, projectRecordID,
	)
	catalogReplay := catalogDemoProvision(
		t, ctx, coordinator, catalogOwner, catalogRequest,
	)
	assertCatalogDemoExactReadyReplay(t, catalogReady, catalogReplay)
	replayEvidence := environment.AssertCatalogDemoCrossPackageEvidence(
		t, ctx, catalogDemoSalesPackageID, catalogDemoCrossLegacy, projectRecordID,
	)
	if !reflect.DeepEqual(evidence, replayEvidence) {
		t.Fatalf("cross-package replay duplicated evidence: before=%+v after=%+v",
			evidence, replayEvidence)
	}
}
