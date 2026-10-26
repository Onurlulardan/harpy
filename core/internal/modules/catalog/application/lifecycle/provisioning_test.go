package lifecycle

import (
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
)

func TestProvisionReceiptRequiresExactPinnedIdentityAndCanonicalRetry(t *testing.T) {
	completedAt := time.Date(2026, 7, 30, 18, 0, 0, 123000, time.UTC)
	request := componentowner.ProvisionRequest{
		PackageID:      "018f0000-0000-7000-8000-000000000801",
		PackageSlug:    "hr",
		PackageVersion: "1.0.0",
		ReleaseID:      "018f0000-0000-7000-8000-000000000802",
		Component: componentowner.Component{
			ID:    "018f0000-0000-7000-8000-000000000803",
			Owner: "forms", Generation: 7,
			Contract:      "forms.catalog-definition.v1",
			ContentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		Document:       []byte(`{"contract_version":"forms.catalog-definition.v1"}`),
		IdempotencyKey: "provision-test",
	}
	ready := componentowner.ProvisionReceipt{
		ID:        "018f0000-0000-7000-8000-000000000804",
		PackageID: request.PackageID, PackageSlug: request.PackageSlug,
		PackageVersion: request.PackageVersion,
		ReleaseID:      request.ReleaseID, Component: request.Component,
		State: componentowner.ProvisionStaged, Ready: true,
		CompletedAt: completedAt,
	}
	if !provisionReceiptMatches(ready, request) {
		t.Fatal("exact ready provision receipt was rejected")
	}
	pending := ready
	pending.Ready = false
	pending.State = componentowner.ProvisionAvailable
	pending.RetryAt = completedAt.Add(time.Second)
	if !provisionReceiptMatches(pending, request) {
		t.Fatal("canonical pending provision receipt was rejected")
	}
	pending.Component.Generation++
	if provisionReceiptMatches(pending, request) {
		t.Fatal("foreign owner generation was accepted")
	}
	pending = ready
	pending.Ready = false
	pending.RetryAt = completedAt.Add(time.Nanosecond)
	if provisionReceiptMatches(pending, request) {
		t.Fatal("noncanonical retry time was accepted")
	}
}

func TestClassifyProvisionResultKeepsOwnerError(t *testing.T) {
	request := componentowner.ProvisionRequest{
		PackageID: "018f0000-0000-7000-8000-000000000801",
	}
	cause := errors.New("project analytics catalog component")
	if err := classifyProvisionResult(cause, componentowner.ProvisionReceipt{}, request); !errors.Is(err, cause) {
		t.Fatalf("owner provision error was remapped: %v", err)
	}
	if err := classifyProvisionResult(nil, componentowner.ProvisionReceipt{}, request); !errors.Is(err, ErrOwnerReceiptStale) {
		t.Fatalf("mismatched receipt was accepted: %v", err)
	}
}

func TestComponentTargetRequiresExactIdentityAndPositiveGeneration(t *testing.T) {
	identity := componentowner.Component{
		ID:    "018f0000-0000-7000-8000-000000000811",
		Owner: "forms", Contract: "forms.catalog-definition.v1",
		ContentSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	target := identity
	target.Generation = 3
	if !componentTargetMatches(target, identity) {
		t.Fatal("exact component target was rejected")
	}
	target.Contract = "forms.catalog-definition.v2"
	if componentTargetMatches(target, identity) {
		t.Fatal("foreign component target was accepted")
	}
}
