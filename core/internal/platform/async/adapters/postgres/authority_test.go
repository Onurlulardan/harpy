package postgres

import (
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

const authorityTenant = "11111111-1111-4111-8111-111111111111"

func TestAuthorityStampMatchesExactJobAndEventAuthority(t *testing.T) {
	t.Parallel()
	stamp := authorityFixture()
	tenant := authorityTenant
	job := jobsv1.Enqueue{
		TenantID: &tenant,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "operations", GroupGeneration: 7,
			Domain: "workflow", CutoverGeneration: 5, WriterOrigin: jobsv1.WriterV2,
		},
	}
	event := eventsv1.Envelope{
		TenantID: &tenant,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "operations", GroupGeneration: 7,
			Domain: "workflow", CutoverGeneration: 5, WriterOrigin: eventsv1.WriterV2,
		},
	}
	if !stamp.matchesJob(job) || !stamp.matchesEvent(event) {
		t.Fatal("exact authority did not match")
	}
	job.Authority.GroupGeneration++
	event.ProducerAuthority.WriterOrigin = eventsv1.WriterV1
	if stamp.matchesJob(job) || stamp.matchesEvent(event) {
		t.Fatal("forged authority matched bound stamp")
	}
}

func TestAuthorityStampRejectsGlobalAndCrossTenantInputs(t *testing.T) {
	t.Parallel()
	stamp := authorityFixture()
	other := "22222222-2222-4222-8222-222222222222"
	job := jobsv1.Enqueue{
		TenantID: &other,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeGlobal, ScopeID: authorityTenant,
			GroupID: "operations", GroupGeneration: 7,
			Domain: "workflow", CutoverGeneration: 5, WriterOrigin: jobsv1.WriterV2,
		},
	}
	event := eventsv1.Envelope{
		TenantID: &other,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: authorityTenant,
			GroupID: "operations", GroupGeneration: 7,
			Domain: "workflow", CutoverGeneration: 5, WriterOrigin: eventsv1.WriterV2,
		},
	}
	if stamp.matchesJob(job) || stamp.matchesEvent(event) {
		t.Fatal("global or cross-tenant input matched tenant authority")
	}
}

func authorityFixture() authorityStamp {
	return authorityStamp{
		scopeID: uuid.MustParse(authorityTenant), groupID: "operations",
		groupGeneration: 7, domain: "workflow", generation: 5,
	}
}
