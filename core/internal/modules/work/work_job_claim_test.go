package work

import (
	"testing"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/api/jobcontract"
	"github.com/google/uuid"
)

type databaseClaimFixture struct {
	jobID, attemptID, tenantID uuid.UUID
	kind                       string
	reference, digest          string
	revision                   int
	generation                 int64
	tenantScoped               bool
}

func (claim databaseClaimFixture) JobID() uuid.UUID     { return claim.jobID }
func (claim databaseClaimFixture) AttemptID() uuid.UUID { return claim.attemptID }
func (claim databaseClaimFixture) TenantID() (uuid.UUID, bool) {
	return claim.tenantID, claim.tenantScoped
}
func (claim databaseClaimFixture) Kind() string             { return claim.kind }
func (claim databaseClaimFixture) PayloadRevision() int     { return claim.revision }
func (claim databaseClaimFixture) PayloadReference() string { return claim.reference }
func (claim databaseClaimFixture) PayloadSHA256() string    { return claim.digest }
func (claim databaseClaimFixture) LeaseGeneration() int64   { return claim.generation }

func TestParseWorkJobClaimPinsKindTenantReferenceAndDigest(t *testing.T) {
	claim := validDatabaseClaim(t, asyncmanifestv1.WorkPurgeExecuteJobKind)
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	parsed, err := parseWorkJobClaim(claim, now)
	if err != nil || !parsed.Valid() || parsed.Kind != claim.kind || parsed.Now != now {
		t.Fatalf("parse claim = %#v, %v", parsed, err)
	}
	for name, mutate := range map[string]func(*databaseClaimFixture){
		"digest": func(value *databaseClaimFixture) { value.digest = "00" + value.digest[2:] },
		"tenant": func(value *databaseClaimFixture) {
			value.tenantID = uuid.MustParse("019b2ee0-1504-7a1a-8d21-5ca5797f2804")
		},
		"kind": func(value *databaseClaimFixture) { value.kind = asyncmanifestv1.WorkProjectionRepairJobKind },
	} {
		t.Run(name, func(t *testing.T) {
			invalid := claim
			mutate(&invalid)
			if _, err := parseWorkJobClaim(invalid, now); err == nil {
				t.Fatal("confused claim was accepted")
			}
		})
	}
}

func validDatabaseClaim(t *testing.T, kind string) databaseClaimFixture {
	t.Helper()
	tenant := uuid.MustParse("019b2ee0-1501-7a1a-8d21-5ca5797f2801")
	target := uuid.MustParse("019b2ee0-1502-7a1a-8d21-5ca5797f2802")
	reference, digest, err := jobcontract.Build(kind, tenant.String(), target.String())
	if err != nil {
		t.Fatal(err)
	}
	return databaseClaimFixture{jobID: uuid.MustParse("019b2ee0-1500-7a1a-8d21-5ca5797f2800"),
		attemptID: uuid.MustParse("019b2ee0-1503-7a1a-8d21-5ca5797f2803"),
		tenantID:  tenant, tenantScoped: true, kind: kind, reference: reference,
		digest: digest, revision: 1, generation: 2}
}
