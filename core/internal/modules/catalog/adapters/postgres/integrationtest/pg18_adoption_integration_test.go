//go:build integration

package integrationtest

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/adoption"
	registrypostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/registry"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type catalogAdoptionAuthorizer struct{}

func (catalogAdoptionAuthorizer) Authorize(
	context.Context, platformtx.TenantScope, string,
) error {
	return nil
}

type catalogAdoptionClock struct{ now time.Time }

func (clock catalogAdoptionClock) Now() time.Time { return clock.now }

func assertCatalogAdoption(t *testing.T, ctx context.Context,
	coordinator *platformtx.Coordinator,
) {
	t.Helper()
	clock := catalogAdoptionClock{
		now: time.Date(2026, 7, 18, 19, 0, 0, 123000, time.UTC),
	}
	application, err := adoptionapp.NewApplication(
		adoptionpostgres.NewBinder(), registrypostgres.NewBinder(),
		catalogAdoptionAuthorizer{}, clock,
	)
	if err != nil {
		t.Fatal(err)
	}
	verifiedRequest := adoptionapp.VerifyRequest{
		RunID:                uuid.MustParse("018f2000-0000-7000-8000-000000000041"),
		SourceManifestSHA256: "4444444444444444444444444444444444444444444444444444444444444444",
		TransformSHA256:      "5555555555555555555555555555555555555555555555555555555555555555",
		CutoverGeneration:    7,
	}
	verified := validateCatalogAdoption(
		t, ctx, coordinator, catalogTenantA, application, verifiedRequest,
	)
	if verified.State != "VERIFIED" || verified.BlockerCode != "" || verified.VerifiedAt == nil {
		t.Fatalf("Catalog adoption verified result=%+v", verified)
	}
	replayed := validateCatalogAdoption(
		t, ctx, coordinator, catalogTenantA, application, verifiedRequest,
	)
	if !reflect.DeepEqual(replayed, verified) {
		t.Fatalf("Catalog adoption replay drifted: %+v / %+v", verified, replayed)
	}

	blockedRequest := verifiedRequest
	blockedRequest.RunID = uuid.MustParse("018f2000-0000-7000-8000-000000000042")
	blockedRequest.CutoverGeneration = 9
	blocked := validateCatalogAdoption(
		t, ctx, coordinator, catalogTenantB, application, blockedRequest,
	)
	wantVersion := "UNKNOWN_VERSION:" + catalogPackageID
	wantManifest := "UNKNOWN_MANIFEST_VERSION:" + catalogLegacySchemaVersionID(catalogTenantB)
	if blocked.State != "BLOCKED" || !strings.Contains(blocked.BlockerCode, wantVersion) ||
		!strings.Contains(blocked.BlockerCode, wantManifest) || blocked.VerifiedAt != nil {
		t.Fatalf("Catalog adoption blocked result=%+v", blocked)
	}
}

func validateCatalogAdoption(t *testing.T, ctx context.Context,
	coordinator *platformtx.Coordinator, tenantID string,
	application *adoptionapp.Application, request adoptionapp.VerifyRequest,
) adoptionapp.Run {
	t.Helper()
	var result adoptionapp.Run
	err := coordinator.InTenant(ctx, catalogAdoptionRequest(t, tenantID), func(
		scoped context.Context, scope platformtx.TenantScope,
	) error {
		var validateErr error
		result, validateErr = application.Validate(scoped, scope, request)
		return validateErr
	})
	if err != nil {
		t.Fatalf("Catalog adoption tenant=%s err=%v", tenantID, err)
	}
	return result
}
