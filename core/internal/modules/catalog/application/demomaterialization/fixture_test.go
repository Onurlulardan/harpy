package demomaterialization

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/google/uuid"
)

const (
	testPackageID = "e3000000-0000-0000-0000-000000000001"
	testReleaseID = "019c0000-0000-7000-8000-000000000001"
	testComponent = "019c0000-0000-7000-8000-000000000002"
	testReceipt   = "019c0000-0000-7000-8000-000000000003"
)

func fixtureRequest(t *testing.T) contract.ProvisionRequest {
	t.Helper()
	document, err := canonicaljsonv1.CanonicalizeObject([]byte(fmt.Sprintf(
		`{"component_id":"%s","contract_version":"%s","demo_forms":[`+
			`{"legacy_form_config_id":"e3000000-0000-0000-0000-000000000002",`+
			`"rows":[{"id":"e3000000-0000-0000-0000-000000000003"}],`+
			`"slug":"employee"}],"excluded_form_slugs":[],"module_slug":`+
			`"human-resources","source_contract":`+
			`"harpy-v1.enterprise-seed.human-resources"}`,
		testComponent, catalogpackagev1.ContractCatalog,
	)), catalogpackagev1.MaximumPayloadBytes)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(document)
	return contract.ProvisionRequest{
		PackageID: testPackageID, PackageSlug: "human-resources",
		PackageVersion: "1.0.0", ReleaseID: testReleaseID,
		Component: contract.Component{
			ID: testComponent, Owner: owner, Generation: 88,
			Contract: catalogContract(), ContentSHA256: hex.EncodeToString(digest[:]),
		},
		Document: document, IdempotencyKey: "catalog-demo-proof",
	}
}

func fixtureSnapshot(
	t *testing.T,
	request contract.ProvisionRequest,
	ready bool,
) (Snapshot, verifiedRequest, [32]byte) {
	t.Helper()
	verified, err := verifyRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	requestSHA, err := projectionDigest(verified)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 30, 14, 40, 0, 123000, time.UTC)
	retryAt := now.Add(30 * time.Second)
	if ready {
		retryAt = time.Time{}
	}
	return Snapshot{
		ReceiptID: uuid.MustParse(testReceipt), PackageID: verified.packageID,
		PackageSlug: request.PackageSlug, PackageVersion: request.PackageVersion,
		ReleaseID: verified.releaseID, Component: request.Component,
		Materialization: verified.materialization.Clone(),
		IdempotencyKey:  request.IdempotencyKey, RequestSHA256: requestSHA,
		State: contract.ProvisionAvailable, Ready: ready, RetryAt: retryAt,
		CompletedAt: now, CreatedAt: now, UpdatedAt: now,
	}, verified, requestSHA
}
