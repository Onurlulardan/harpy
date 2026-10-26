package iam

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestIdentityAuthorizationFailsClosedBeforeRepositoryUse(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	application := &identityHTTPApplication{}
	if _, err := application.authorize(
		context.Background(), platformtx.TenantScope{}, fixture.claims,
		"unknown", "read",
	); !errors.Is(err, ErrIdentityHTTPInvalid) {
		t.Fatalf("unknown resource error=%v", err)
	}
	if _, err := application.authorize(
		context.Background(), platformtx.TenantScope{}, fixture.claims,
		IdentityUsers, "read",
	); !errors.Is(err, ErrIdentityHTTPUnavailable) {
		t.Fatalf("unavailable evidence error=%v", err)
	}
	if _, err := application.authorize(
		context.Background(), platformtx.TenantScope{}, accesstoken.Claims{},
		IdentityUsers, "read",
	); !errors.Is(err, ErrIdentityHTTPForbidden) {
		t.Fatalf("invalid claim versions error=%v", err)
	}
}

func TestIdentityAuthorizationEvidenceBindsActorVersionsAndPermission(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	permissionID := uuid.MustParse("78000000-0000-4000-8000-000000000001")
	operationID := uuid.MustParse("79000000-0000-8000-8000-000000000001")
	digest := [32]byte{1, 2, 3}
	query := identityQueryEvidence(fixture.claims, permissionID)
	mutation := identityMutationEvidence(fixture.claims, permissionID, operationID, digest)
	if uuid.UUID(query.ActorID) != fixture.claims.Subject() ||
		uuid.UUID(query.PermissionID) != permissionID ||
		query.ExpectedTokenVersion != fixture.claims.TokenVersion() ||
		query.ExpectedAuthorizationVersion != fixture.claims.AuthorizationVersion() {
		t.Fatalf("query evidence=%+v", query)
	}
	if uuid.UUID(mutation.OperationID) != operationID || mutation.RequestSHA256 != digest ||
		uuid.UUID(mutation.ActorID) != fixture.claims.Subject() ||
		uuid.UUID(mutation.PermissionID) != permissionID ||
		mutation.ExpectedTokenVersion != fixture.claims.TokenVersion() ||
		mutation.ExpectedAuthorizationVersion != fixture.claims.AuthorizationVersion() {
		t.Fatalf("mutation evidence=%+v", mutation)
	}
}
