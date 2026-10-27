package publication

import (
	"context"
	"errors"
	"testing"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
)

func TestExecuteUsesOpaqueRouteAndCapabilityHash(t *testing.T) {
	lease := &publicationFixtureLease{}
	limiter := &publicationFixtureLimiter{lease: lease}
	scopes := &publicationFixtureScopes{}
	executor := &publicationFixtureExecutor{result: PublicResult{Body: []byte(`{"ok":true}`)}}
	application, err := NewPublicApplication(
		publicationFixtureDirectory{route: publicationFixtureRoute(analyticspublication.AudienceCapability)},
		scopes, executor, limiter, publicationFixtureClock{publicationFixtureNow})
	if err != nil {
		t.Fatal(err)
	}
	result, err := application.Execute(context.Background(), PublicRequest{
		PublicID: publicationFixturePublicID, ExpectedKind: ResourceQuery,
		Slug: "cosmetic-only", TrustedClientIP: "203.0.113.10",
		CapabilityToken: publicationFixtureToken, Parameters: []byte(`{}`)})
	if err != nil || string(result.Body) != `{"ok":true}` || !limiter.called ||
		limiter.admission.TenantID != publicationFixtureTenant || !scopes.called ||
		!executor.called || !lease.released {
		t.Fatalf("public execute = %s, limiter=%v scopes=%v executor=%v release=%v err=%v",
			result.Body, limiter.called, scopes.called, executor.called, lease.released, err)
	}
}

func TestExecuteRejectsWrongCapabilityBeforeLimiter(t *testing.T) {
	limiter := &publicationFixtureLimiter{lease: &publicationFixtureLease{}}
	application, _ := NewPublicApplication(
		publicationFixtureDirectory{route: publicationFixtureRoute(analyticspublication.AudienceCapability)},
		&publicationFixtureScopes{}, &publicationFixtureExecutor{}, limiter,
		publicationFixtureClock{publicationFixtureNow})
	_, err := application.Execute(context.Background(), PublicRequest{PublicID: publicationFixturePublicID,
		ExpectedKind: ResourceQuery, Slug: "opaque", TrustedClientIP: "203.0.113.10",
		CapabilityToken: make([]byte, 32),
		Parameters:      []byte(`{}`)})
	if !errors.Is(err, ErrForbidden) || limiter.called {
		t.Fatalf("capability error = %v", err)
	}
}

func TestExecuteRejectsPublicResourceKindSubstitution(t *testing.T) {
	limiter := &publicationFixtureLimiter{lease: &publicationFixtureLease{}}
	application, _ := NewPublicApplication(
		publicationFixtureDirectory{route: publicationFixtureRoute(analyticspublication.AudiencePublic)},
		&publicationFixtureScopes{}, &publicationFixtureExecutor{}, limiter,
		publicationFixtureClock{publicationFixtureNow})
	_, err := application.Execute(context.Background(), PublicRequest{
		PublicID: publicationFixturePublicID, ExpectedKind: ResourceDashboard,
		Slug: "opaque", TrustedClientIP: "203.0.113.10", Parameters: []byte(`{}`)})
	if !errors.Is(err, ErrForbidden) || limiter.called {
		t.Fatalf("kind substitution error = %v", err)
	}
}
