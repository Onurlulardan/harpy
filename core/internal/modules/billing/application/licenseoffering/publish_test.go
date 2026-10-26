package licenseoffering

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	licensebootstrapv1 "github.com/Onurlulardan/harpy/contracts/licensebootstrap/v1"
	licensebootstrap "github.com/Onurlulardan/harpy/core/internal/capabilities/licensebootstrap/api"
)

type offeringStore struct {
	calls   int
	request PublishRequest
}

func (store *offeringStore) PublishLicenseOffering(
	_ context.Context, request PublishRequest,
) (Receipt, error) {
	store.calls++
	store.request = request
	return Receipt{
		OperationID:       request.Binding.OperationID,
		OfferingVersionID: request.Binding.OfferingVersionID,
		Revision:          request.Binding.Revision,
		OfferingSHA256:    request.Binding.OfferingSHA256,
	}, nil
}

func TestPublishReverifiesAndPublishesExactImmutableOffering(t *testing.T) {
	fixture := newOfferingFixture(t)
	grant := offeringGrant(t, fixture)
	store := &offeringStore{}
	publisher, err := NewPublisher(store, fixture.checker())
	if err != nil {
		t.Fatal(err)
	}
	err = licensebootstrap.Run(grant, fixture.now, func(scope licensebootstrap.Scope) error {
		_, publishErr := publisher.Publish(
			context.Background(), scope, fixture.document, fixture.now,
		)
		return publishErr
	})
	if err != nil || store.calls != 1 ||
		store.request.Offering.SHA256 != grant.Binding.OfferingSHA256 ||
		store.request.Binding.Sequence != grant.Binding.Sequence {
		t.Fatalf("offering publication failed: calls=%d err=%v", store.calls, err)
	}
}

func TestPublishRejectsTamperAndWrongPlannedProjection(t *testing.T) {
	fixture := newOfferingFixture(t)
	store := &offeringStore{}
	publisher, _ := NewPublisher(store, fixture.checker())
	tampered := append([]byte(nil), fixture.document...)
	tampered[len(tampered)-5] ^= 1
	err := licensebootstrap.Run(
		offeringGrant(t, fixture), fixture.now, func(scope licensebootstrap.Scope) error {
			_, publishErr := publisher.Publish(
				context.Background(), scope, tampered, fixture.now,
			)
			return publishErr
		})
	if !errors.Is(err, ErrBinding) || store.calls != 0 {
		t.Fatalf("tampered snapshot published: calls=%d err=%v", store.calls, err)
	}
	grant := offeringGrant(t, fixture)
	grant.Binding.OfferingSHA256 = strings.Repeat("f", 64)
	err = licensebootstrap.Run(grant, fixture.now, func(scope licensebootstrap.Scope) error {
		_, publishErr := publisher.Publish(context.Background(), scope, fixture.document, fixture.now)
		return publishErr
	})
	if !errors.Is(err, ErrProjection) || store.calls != 0 {
		t.Fatalf("wrong planned projection published: calls=%d err=%v", store.calls, err)
	}
}

func offeringGrant(t *testing.T, fixture offeringFixture) licensebootstrap.Grant {
	t.Helper()
	snapshot := fixture.snapshot
	document, err := BuildDocument(Plan{
		OfferingVersionID: fixture.offeringVersionID, Revision: 1, EffectiveAt: fixture.now,
	}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return licensebootstrap.Grant{
		Revision: 1, LeaseUntil: fixture.now.Add(time.Minute),
		Binding: licensebootstrap.Binding{
			OperationID: mustV7(t), EnvironmentID: fixture.environmentID,
			LicenseID: fixture.licenseID, LicenseeOrganizationID: fixture.organizationID,
			DeploymentMode: "onpremise", DeliveryKind: "connected", Issuer: snapshot.Issuer(),
			Sequence: snapshot.Sequence(), SnapshotSHA256: snapshot.Digest(),
			OfferingVersionID: fixture.offeringVersionID, Revision: 1,
			OfferingSHA256: document.SHA256, EffectiveAt: fixture.now,
			ManifestSHA256: licensebootstrapv1.ManifestSHA256(),
		},
	}
}
