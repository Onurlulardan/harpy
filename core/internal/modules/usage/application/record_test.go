package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func TestRecordPassesCanonicalFactAndDigestToStore(t *testing.T) {
	called := false
	store := &stubStore{record: func(value domain.Observation, digest [32]byte) (RecordResult, error) {
		called = true
		if value.ID != domain.ID(testFact) || value.Method != "GET" ||
			value.RoutePattern != "/api/forms/{id}" || digest != value.Digest() {
			t.Fatal("record was not canonical or digest-bound")
		}
		return RecordResult{FactID: uuid.UUID(value.ID), Replayed: true}, nil
	}}
	service, _ := NewService(store, fixedClock{testNow}, fixedIDs{testFact})
	var source [32]byte
	source[0] = 1
	result, err := service.Record(context.Background(), RecordRequest{
		EnvironmentID: testEnvironment, TenantID: testTenant, ActorID: testActor,
		SourceIdentity: source, Method: " get ", RoutePattern: " /api/forms/{id} ",
		StatusCode: 200, DurationMicros: 1500, OccurredAt: testNow,
	})
	if err != nil || !called || !result.Replayed || result.FactID != testFact {
		t.Fatalf("record failed: result=%+v err=%v", result, err)
	}
}

func TestRecordPropagatesSourceIdentityConflict(t *testing.T) {
	store := &stubStore{record: func(domain.Observation, [32]byte) (RecordResult, error) {
		return RecordResult{}, domain.ErrConflict
	}}
	service, _ := NewService(store, fixedClock{testNow}, fixedIDs{testFact})
	var source [32]byte
	source[0] = 1
	_, err := service.Record(context.Background(), RecordRequest{
		EnvironmentID: testEnvironment, TenantID: testTenant, ActorID: testActor,
		SourceIdentity: source, Method: "POST", RoutePattern: "/api/forms",
		StatusCode: 201, OccurredAt: testNow,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("conflict was not propagated: %v", err)
	}
}

func TestObservationDigestDoesNotDependOnGeneratedTargetID(t *testing.T) {
	var source [32]byte
	source[0] = 1
	base := domain.Observation{ID: domain.ID(testFact),
		EnvironmentID: domain.ID(testEnvironment), TenantID: domain.ID(testTenant),
		ActorID: domain.ID(testActor), SourceIdentity: source,
		Method: "GET", RoutePattern: "/api/forms/{id}", StatusCode: 200,
		OccurredAt: testNow}
	other := base
	other.ID = domain.ID(uuid.MustParse("01890f9e-7b6c-7cc7-98c4-dc0c0c073990"))
	if base.Digest() != other.Digest() {
		t.Fatal("source payload digest depends on generated target identity")
	}
}

func TestRecordRejectsRawQueryPath(t *testing.T) {
	service := testService()
	var source [32]byte
	source[0] = 1
	_, err := service.Record(context.Background(), RecordRequest{
		EnvironmentID: testEnvironment, TenantID: testTenant, ActorID: testActor,
		SourceIdentity: source, Method: "GET", RoutePattern: "/api/forms?id=secret",
		StatusCode: 200, OccurredAt: testNow,
	})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("raw query path accepted: %v", err)
	}
}
