package api

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type catalogIAMAccessFixture struct {
	decision iammodule.CatalogAccessDecision
	err      error
	request  iammodule.CatalogAccessRequest
}

func (fixture *catalogIAMAccessFixture) Verify(
	_ context.Context, request iammodule.CatalogAccessRequest,
) (iammodule.CatalogAccessDecision, error) {
	fixture.request = request
	return fixture.decision, fixture.err
}

func TestCatalogAccessVerifierPreservesIAMDecisionAndPurpose(t *testing.T) {
	principalID := [16]byte{0x10, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	tenantID := [16]byte{0x20, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	stepUp := time.Date(2026, 7, 19, 8, 9, 10, 0, time.UTC)
	permissionID := [16]byte{0x30, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	authorizationSHA := sha256.Sum256([]byte("catalog authorization"))
	purposeSHA := sha256.Sum256([]byte("catalog purge purpose"))
	decision, err := iammodule.NewCatalogAccessDecisionWithEvidence(
		principalID, tenantID, 3, 5, 7, stepUp, permissionID,
		authorizationSHA, purposeSHA,
	)
	if err != nil {
		t.Fatal(err)
	}
	iam := &catalogIAMAccessFixture{decision: decision}
	verifier, err := newCatalogAccessVerifier(iam)
	if err != nil {
		t.Fatal(err)
	}
	request := catalogmodule.HTTPAccessRequest{
		AuthorizationValues: []string{"Bearer token"},
		TenantIDValues:      []string{decision.TenantID().String()},
		TenantSlugValues:    []string{"tenant"},
		Permission:          catalogmodule.HTTPPermission("catalog.purge"),
		Operation:           "catalog.installation.purge",
		PurposeTarget:       "catalog-package:sample",
	}
	access, err := verifier.Verify(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if access.PrincipalID() != principalID || access.TenantID() != tenantID ||
		access.TokenVersion() != 3 || access.AuthorizationVersion() != 5 ||
		access.AuthorizationClock() != 7 || access.StepUpAt() != stepUp ||
		access.AuthorizationPermissionID() != permissionID ||
		access.AuthorizationEvidenceSHA256() != authorizationSHA ||
		access.StepUpPurposeSHA256() != purposeSHA {
		t.Fatalf("Catalog access decision drifted: %#v", access)
	}
	if iam.request.Permission != "catalog.purge" ||
		iam.request.Operation != request.Operation ||
		iam.request.PurposeTarget != request.PurposeTarget {
		t.Fatalf("Catalog IAM request drifted: %#v", iam.request)
	}
}

func TestCatalogAccessVerifierMapsIAMFailuresFailClosed(t *testing.T) {
	tests := []struct {
		name string
		in   error
		want error
	}{
		{"authentication", iammodule.ErrCatalogAccessAuthentication,
			catalogmodule.ErrHTTPAuthentication},
		{"forbidden", iammodule.ErrCatalogAccessForbidden, catalogmodule.ErrHTTPForbidden},
		{"unavailable", iammodule.ErrCatalogAccessUnavailable,
			catalogmodule.ErrHTTPUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifier, err := newCatalogAccessVerifier(&catalogIAMAccessFixture{err: test.in})
			if err != nil {
				t.Fatal(err)
			}
			_, err = verifier.Verify(context.Background(), catalogmodule.HTTPAccessRequest{})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
