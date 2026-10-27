//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type credentialIntegrationResolver struct {
	database   postgres.BoundExecutor
	tenantID   uuid.UUID
	generation int64
}

func (resolver credentialIntegrationResolver) CredentialTarget(
	ctx context.Context,
	lookup CredentialLookup,
) (CredentialTargetEvidence, error) {
	if NetworkAllowed(ctx) {
		return CredentialTargetEvidence{}, errors.New("resolver network boundary is open")
	}
	var probe int
	if err := resolver.database.QueryRow(ctx, "SELECT 1").Scan(&probe); err != nil || probe != 1 {
		return CredentialTargetEvidence{}, errors.New("resolver transaction is unavailable")
	}
	evidence := CredentialTargetEvidence{
		OperationID: lookup.OperationID(), TenantID: resolver.tenantID,
		Purpose: lookup.Purpose(), LocatorKind: lookup.LocatorKind(),
		LocatorBindingSHA256: [32]byte{1}, SourceRevision: 1,
		GroupGeneration: resolver.generation, IAMGeneration: resolver.generation,
		SystemResourceGeneration: resolver.generation,
	}
	switch lookup.LocatorKind() {
	case CredentialLoginDirectory:
		evidence.SourceUserID = uuid.New()
		evidence.SourceIdentifierID = uuid.New()
		evidence.SourceIdentifierKind = lookup.IdentifierKind()
	case CredentialRefreshLocator:
		evidence.SourceTokenID = uuid.New()
		evidence.SourceFamilyID = uuid.New()
		copy(evidence.SourceSelector[:], lookup.Key())
	}
	return evidence, nil
}

func assertCredentialTenantScope(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	request := credentialScopeRequestFor(
		t, tenantA, 1, CredentialLogin, CredentialLoginDirectory,
	)
	binder := scopeProbeBinder(t)
	rowID := uuid.New()
	var retained CredentialTenantScope
	err := coordinator.InCredentialTenant(
		ctx, request,
		func(scoped context.Context, scope CredentialTenantScope) error {
			retained = scope
			if !ContextOwnsCredentialScope(scoped, scope) || NetworkAllowed(scoped) {
				t.Fatal("credential callback escaped its context boundary")
			}
			if _, err := MintOwnerAuthority(TenantScope{}, "iam"); !errors.Is(err, ErrScopeClosed) {
				t.Fatalf("credential path minted owner authority: %v", err)
			}
			probe, err := BindCredential(scope, binder)
			if err != nil {
				return err
			}
			_, err = probe.database.Execute(
				scoped, "INSERT INTO iam_probe(id,value) VALUES($1,$2)", rowID, "credential",
			)
			return err
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if retained.valid() {
		t.Fatal("credential scope remained active after commit")
	}
	assertCredentialRowCommitted(t, ctx, coordinator, rowID)
	stale := credentialScopeRequestFor(
		t, tenantA, 2, CredentialLogin, CredentialLoginDirectory,
	)
	called := false
	err = coordinator.InCredentialTenant(
		ctx, stale, func(context.Context, CredentialTenantScope) error {
			called = true
			return nil
		},
	)
	if !errors.Is(err, ErrWriterAuthority) || called {
		t.Fatalf("stale authority result: called=%v err=%v", called, err)
	}
}

func credentialScopeRequestFor(
	t *testing.T,
	tenant string,
	generation int64,
	purpose CredentialPurpose,
	locator CredentialLocatorKind,
) CredentialScopeRequest {
	t.Helper()
	lookup, err := credentialLookupForIntegration(uuid.MustParse(tenant), purpose, locator)
	if err != nil {
		t.Fatal(err)
	}
	binder, err := NewCredentialResolverBinder(
		func(database postgres.BoundExecutor) CredentialTargetResolver {
			return credentialIntegrationResolver{
				database: database, tenantID: uuid.MustParse(tenant), generation: generation,
			}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return CredentialScopeRequest{Lookup: lookup, Resolver: binder}
}

func credentialLookupForIntegration(
	tenantID uuid.UUID,
	purpose CredentialPurpose,
	locator CredentialLocatorKind,
) (CredentialLookup, error) {
	const operationID = "iam.credential.integration"
	switch locator {
	case CredentialLoginDirectory:
		return NewLoginCredentialLookup(
			operationID, tenantID,
			CredentialEmailIdentifier, testCredentialKeyID, [32]byte{1},
		)
	case CredentialRefreshLocator:
		return NewRefreshCredentialLookup(operationID, [16]byte{1})
	case CredentialSessionLocator:
		return NewSessionCredentialLookup(operationID, tenantID, [32]byte{1})
	default:
		return NewTenantRouteCredentialLookup(
			operationID, purpose, tenantID, [32]byte{1},
		)
	}
}

func assertCredentialRowCommitted(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	rowID uuid.UUID,
) {
	t.Helper()
	err := coordinator.InTenant(
		ctx, scopeRequestFor(t, tenantA, 1),
		func(ctx context.Context, scope TenantScope) error {
			probe, err := Bind(scope, scopeProbeBinder(t))
			if err != nil {
				return err
			}
			var value string
			err = probe.database.QueryRow(
				ctx, "SELECT value FROM iam_probe WHERE id=$1", rowID,
			).Scan(&value)
			if err != nil {
				return err
			}
			if value != "credential" {
				t.Fatalf("credential row value = %q", value)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}
