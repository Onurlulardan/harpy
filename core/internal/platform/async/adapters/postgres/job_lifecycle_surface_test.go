package postgres

import (
	"reflect"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func TestJobLifecycleAuthoritySurfacesAreFailClosed(t *testing.T) {
	t.Parallel()
	worker := reflect.TypeOf((*WorkerStore)(nil)).Elem()
	for _, forbidden := range []string{
		"CompleteDatabaseJob", "MarkExternalEffectStarted",
		"FailJob", "FailExternalJobBeforeEffect",
	} {
		if _, exposed := worker.MethodByName(forbidden); exposed {
			t.Fatalf("WorkerStore exposes tenant-bound method %s", forbidden)
		}
	}
	writer := reflect.TypeOf((*Writer)(nil)).Elem()
	for _, required := range []string{
		"CompleteDatabaseJob", "MarkExternalEffectStarted", "FailJobBeforeEffect",
	} {
		if _, exposed := writer.MethodByName(required); !exposed {
			t.Fatalf("Writer does not expose bound method %s", required)
		}
	}
}

func TestAuthorityStampMatchesOnlyItsTenantClaim(t *testing.T) {
	t.Parallel()
	repository, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	stamp := authorityStamp{
		scopeID: claim.tenantID, groupID: claim.authority.groupID,
		groupGeneration: claim.authority.groupGeneration,
		domain:          claim.authority.domain, generation: claim.authority.generation,
	}
	if !stamp.matchesClaim(claim) {
		t.Fatal("exact tenant claim did not match its bound authority")
	}
	other := stamp
	other.scopeID = uuid.New()
	if other.matchesClaim(claim) {
		t.Fatal("cross-tenant claim matched bound authority")
	}
	if stamp.matchesClaim(validGlobalClaim(t, repository)) {
		t.Fatal("global claim entered tenant-bound lifecycle")
	}
}
