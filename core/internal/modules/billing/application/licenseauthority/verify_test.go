package licenseauthority

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"
)

func TestVerifyProducesBoundImmutableSnapshot(t *testing.T) {
	fixture := newAuthorityFixture(t)
	snapshot, err := Verify(
		context.Background(), fixture.document(t), fixture.binding, fixture.now,
		fixture.resolver(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.LicenseID() != fixture.binding.LicenseID || snapshot.Sequence() != 7 ||
		snapshot.RuntimeState(fixture.now) != RuntimeValid {
		t.Fatalf("unexpected verified snapshot: %+v", snapshot)
	}
	projection := ProjectEntitlements(snapshot)
	if !projection.Boolean["forms.builder.enabled"] ||
		!projection.Limits["forms.active_forms"].Unlimited ||
		projection.Limits["iam.active_users"].Value != 25 {
		t.Fatalf("unexpected entitlement projection: %+v", projection)
	}
	if snapshot.RuntimeState(snapshot.LicenseGrantDeadline().Add(time.Second)) != RuntimeGrace ||
		snapshot.RuntimeState(snapshot.GraceUntil().Add(time.Second)) != RuntimeBuildOff {
		t.Fatal("runtime license clock state drifted")
	}
}

func TestVerifyRejectsTamperBindingAndFutureSnapshot(t *testing.T) {
	fixture := newAuthorityFixture(t)
	document := fixture.document(t)
	document[len(document)-8] ^= 1
	if _, err := Verify(
		context.Background(), document, fixture.binding, fixture.now, fixture.resolver(),
	); !errors.Is(err, ErrInvalidSignature) && !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("tamper accepted: %v", err)
	}

	other := fixture.binding
	other.EnvironmentID[0] ^= 1
	if _, err := Verify(
		context.Background(), fixture.document(t), other, fixture.now, fixture.resolver(),
	); !errors.Is(err, ErrWrongBinding) {
		t.Fatalf("wrong binding accepted: %v", err)
	}

	fixture.payload.NotBefore = fixture.now.Add(time.Minute).Format(time.RFC3339)
	fixture.payload.RefreshAfter = fixture.now.Add(2 * time.Minute).Format(time.RFC3339)
	if _, err := Verify(
		context.Background(), fixture.document(t), fixture.binding, fixture.now,
		fixture.resolver(),
	); !errors.Is(err, ErrNotYetValid) {
		t.Fatalf("future snapshot accepted: %v", err)
	}
}

func TestVerifyRejectsUnknownClaimsAndWrongTrustKey(t *testing.T) {
	fixture := newAuthorityFixture(t)
	document := fixture.document(t)
	invalid := append([]byte(`{"unknown":true,`), document[1:]...)
	if _, err := Verify(
		context.Background(), invalid, fixture.binding, fixture.now, fixture.resolver(),
	); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("unknown envelope claim accepted: %v", err)
	}
	resolver := KeyResolverFunc(func(
		context.Context, string, string, time.Time, time.Time,
	) (ed25519.PublicKey, error) {
		return nil, ErrUntrustedKey
	})
	if _, err := Verify(
		context.Background(), document, fixture.binding, fixture.now, resolver,
	); !errors.Is(err, ErrUntrustedKey) {
		t.Fatalf("untrusted key accepted: %v", err)
	}
}

func TestVerifyRejectsDuplicateEnvelopeFields(t *testing.T) {
	fixture := newAuthorityFixture(t)
	document := fixture.document(t)
	duplicate := append([]byte(`{"schema":"harpy-license-envelope.v2",`), document[1:]...)
	if _, err := Verify(
		context.Background(), duplicate, fixture.binding, fixture.now, fixture.resolver(),
	); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("duplicate envelope field accepted: %v", err)
	}
}
