package publication

import (
	"context"
	"errors"
	"testing"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExposureWideningDenialPreventsEveryPublicationOwnerWrite(t *testing.T) {
	run := func(
		name string,
		invoke func(*Application, *publicationMutationStore) error,
	) {
		t.Run(name, func(t *testing.T) {
			entitlements := &publicationFixtureEntitlements{
				decision: billingapi.MeteredResourceMutationDecision{
					Resource: billingapi.MeteredResourceAPIEndpoint,
					Mutation: billingapi.MeteredResourcePublish,
				},
			}
			application := publicationMutationApplication(
				entitlements, publicationFixtureOperation, publicationFixtureRelease,
			)
			store := &publicationMutationStore{}
			err := invoke(application, store)
			if !errors.Is(err, ErrEntitlement) {
				t.Fatalf("%s error = %v", name, err)
			}
			if !entitlements.called || store.writes != 0 {
				t.Fatalf("%s Billing called=%v owner writes=%d",
					name, entitlements.called, store.writes)
			}
		})
	}

	identity := publicationMutationIdentity()
	current := publicationMutationCurrent()
	timing := publicationMutationTiming()
	scope := platformtx.TenantScope{}
	run("create", func(application *Application, store *publicationMutationStore) error {
		_, err := application.createBound(context.Background(), scope, store, identity,
			CreateRequest{ResourceKind: ResourceQuery,
				ResourceVersionID: publicationFixtureVersion,
				Audience:          analyticspublication.AudiencePublic,
				CachePolicy:       analyticspublication.CacheNone,
				IdempotencyKey:    "publication-create-denied"}, [32]byte{1})
		return err
	})
	run("submit", func(application *Application, store *publicationMutationStore) error {
		_, err := application.submitBound(context.Background(), scope, store, identity,
			current, timing, TransitionRequest{ReleaseID: publicationFixtureRelease,
				ExpectedRevision: 1, IdempotencyKey: "publication-submit-denied"})
		return err
	})
	run("approve", func(application *Application, store *publicationMutationStore) error {
		_, err := application.approveBound(context.Background(), scope, store, identity,
			current, timing, TransitionRequest{ReleaseID: publicationFixtureRelease,
				ExpectedRevision: 1, IdempotencyKey: "publication-approve-denied"})
		return err
	})
	run("activate", func(application *Application, store *publicationMutationStore) error {
		_, err := application.activateBound(context.Background(), scope, store, identity,
			current, timing, ActivateRequest{ReleaseID: publicationFixtureRelease,
				ExpectedRevision: 1, IdempotencyKey: "publication-activate-denied"})
		return err
	})
}

func TestPublicationCreateOwnerFailureRollsBackBillingDecisionScope(t *testing.T) {
	entitlements := &publicationFixtureEntitlements{}
	application := publicationMutationApplication(entitlements,
		publicationFixtureOperation, publicationFixtureRelease)
	ownerFailure := errors.New("publication owner write failed")
	store := &publicationMutationStore{create: func(
		CreateMutation,
	) (Snapshot, bool, error) {
		if !entitlements.called {
			t.Fatal("publication owner store called before Billing decision")
		}
		return Snapshot{}, false, ownerFailure
	}}
	_, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, publicationMutationIdentity(), CreateRequest{
			ResourceKind: ResourceQuery, ResourceVersionID: publicationFixtureVersion,
			Audience:       analyticspublication.AudiencePublic,
			CachePolicy:    analyticspublication.CacheNone,
			IdempotencyKey: "publication-owner-failure",
		}, [32]byte{1})
	if !errors.Is(err, ownerFailure) || store.writes != 1 {
		t.Fatalf("owner failure = %v, writes = %d", err, store.writes)
	}
}

func TestPublicationMapsBillingPublicMutationErrorsBeforeOwnerWrite(t *testing.T) {
	tests := []struct {
		name    string
		billing error
		want    error
	}{
		{name: "conflict", billing: billingapi.ErrMeteredResourceIdempotencyReused,
			want: ErrConflict},
		{name: "invalid", billing: billingapi.ErrMeteredResourceRequest,
			want: ErrInvalid},
		{name: "unavailable", billing: billingapi.ErrMeteredResourceEvidence,
			want: ErrUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entitlements := &publicationFixtureEntitlements{err: test.billing}
			application := publicationMutationApplication(entitlements,
				publicationFixtureOperation, publicationFixtureRelease)
			store := &publicationMutationStore{}
			_, err := application.createBound(context.Background(), platformtx.TenantScope{},
				store, publicationMutationIdentity(), CreateRequest{
					ResourceKind: ResourceQuery, ResourceVersionID: publicationFixtureVersion,
					Audience:       analyticspublication.AudiencePublic,
					CachePolicy:    analyticspublication.CacheNone,
					IdempotencyKey: "publication-billing-error",
				}, [32]byte{1})
			if !errors.Is(err, test.want) || store.writes != 0 {
				t.Fatalf("mapped error = %v, owner writes = %d", err, store.writes)
			}
		})
	}
}
