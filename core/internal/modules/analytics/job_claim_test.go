package analytics

import (
	"crypto/sha256"
	"fmt"
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/google/uuid"
)

var (
	analyticsJobID      = uuid.MustParse("018f0000-0000-7000-8000-000000000501")
	analyticsAttemptID  = uuid.MustParse("018f0000-0000-7000-8000-000000000502")
	analyticsJobTenant  = uuid.MustParse("018f0000-0000-7000-8000-000000000503")
	analyticsJobQuery   = uuid.MustParse("018f0000-0000-7000-8000-000000000504")
	analyticsJobVersion = uuid.MustParse("018f0000-0000-7000-8000-000000000505")
)

type analyticsJobClaimFixture struct {
	kind, reference, digest string
}

func (claim analyticsJobClaimFixture) JobID() uuid.UUID     { return analyticsJobID }
func (claim analyticsJobClaimFixture) AttemptID() uuid.UUID { return analyticsAttemptID }
func (claim analyticsJobClaimFixture) TenantID() (uuid.UUID, bool) {
	return analyticsJobTenant, true
}
func (claim analyticsJobClaimFixture) Kind() string             { return claim.kind }
func (claim analyticsJobClaimFixture) PayloadRevision() int     { return 1 }
func (claim analyticsJobClaimFixture) PayloadReference() string { return claim.reference }
func (claim analyticsJobClaimFixture) PayloadSHA256() string    { return claim.digest }
func (claim analyticsJobClaimFixture) LeaseGeneration() int64   { return 2 }

func newAnalyticsJobClaim(kind, reference string) analyticsJobClaimFixture {
	digest := sha256.Sum256([]byte(reference))
	return analyticsJobClaimFixture{kind: kind, reference: reference,
		digest: fmt.Sprintf("%x", digest)}
}

func TestAnalyticsQueryJobClaimPinsExactReferenceDigest(t *testing.T) {
	manifest := fmt.Sprintf("%064x", 7)
	reference := fmt.Sprintf("analytics-cache-warm/%s/%s/%s",
		analyticsJobQuery, analyticsJobVersion, manifest)
	claim := newAnalyticsJobClaim(asyncmanifestv1.AnalyticsCacheWarmJobKind, reference)
	parsed, err := parseAnalyticsQueryJobClaim(claim,
		asyncmanifestv1.AnalyticsCacheWarmJobKind, "analytics-cache-warm")
	if err != nil || parsed.queryID != analyticsJobQuery ||
		parsed.versionID != analyticsJobVersion || parsed.manifest == [32]byte{} {
		t.Fatalf("parsed=%+v err=%v", parsed, err)
	}
	claim.digest = fmt.Sprintf("%064x", 8)
	if _, err = parseAnalyticsQueryJobClaim(claim,
		asyncmanifestv1.AnalyticsCacheWarmJobKind, "analytics-cache-warm"); err == nil {
		t.Fatal("tampered Analytics job digest accepted")
	}
}

func TestAnalyticsPublicationJobClaimPinsGenerationAndState(t *testing.T) {
	reference := fmt.Sprintf("analytics-publication-repair/%s/4/revoked",
		analyticsJobVersion)
	claim := newAnalyticsJobClaim(
		asyncmanifestv1.AnalyticsPublicationRepairJobKind, reference)
	parsed, err := parseAnalyticsPublicationJobClaim(
		claim, asyncmanifestv1.AnalyticsPublicationRepairJobKind)
	if err != nil || parsed.releaseID != analyticsJobVersion ||
		parsed.generation != 4 || parsed.state != "revoked" {
		t.Fatalf("parsed=%+v err=%v", parsed, err)
	}
}
