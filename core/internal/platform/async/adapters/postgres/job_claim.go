package postgres

import (
	"regexp"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

var (
	claimDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	claimKindPattern   = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
)

type JobClaim struct {
	authority        JobAuthority
	attemptID        uuid.UUID
	jobID            uuid.UUID
	tenantID         uuid.UUID
	tenantScoped     bool
	kind             string
	payloadRevision  int
	payloadReference string
	payloadSHA256    string
	effectClass      jobsv1.EffectClass
	effectKey        string
	hasEffectKey     bool
	priority         int
	runAt            time.Time
	attempt          int
	maximumAttempts  int
	leaseOwnerID     uuid.UUID
	leaseGeneration  int64
	startedAt        time.Time
	leaseUntil       time.Time
}

func newJobClaim(
	row jobClaimRow,
	authority JobAuthority,
	workerID uuid.UUID,
) (JobClaim, error) {
	tenantID := uuid.Nil
	if row.tenantID.Valid {
		tenantID = uuid.UUID(row.tenantID.Bytes)
	}
	tenantValid := row.scopeKind == "tenant" && row.tenantID.Valid &&
		tenantID == row.scopeID
	globalValid := row.scopeKind == "global" && !row.tenantID.Valid
	effectClass := jobsv1.EffectClass(row.effectClass)
	effectValid := effectClass == jobsv1.EffectDatabaseTransactional &&
		!row.effectKey.Valid || effectClass == jobsv1.EffectExternal &&
		row.effectKey.Valid && len(row.effectKey.String) >= 1 &&
		len(row.effectKey.String) <= jobsv1.MaximumKeyBytes
	if !authority.valid(authority.owner) || row.attemptID == uuid.Nil ||
		row.attemptID.Version() != uuid.Version(7) || row.jobID == uuid.Nil ||
		row.scopeKind != authority.scopeKind || row.scopeID != authority.scopeID ||
		row.groupID != authority.groupID ||
		row.groupGeneration != authority.groupGeneration ||
		row.domain != authority.domain || row.generation != authority.generation ||
		(!tenantValid && !globalValid) || workerID == uuid.Nil ||
		row.leaseOwnerID != workerID || row.attempt < 1 ||
		row.attempt > row.maximumAttempts ||
		row.maximumAttempts > jobsv1.MaximumMaximumAttempts ||
		row.leaseGeneration != int64(row.attempt) || row.startedAt.IsZero() ||
		!row.leaseUntil.Equal(row.startedAt.Add(jobsv1.LeaseSeconds*time.Second)) ||
		row.runAt.After(row.startedAt) || !claimKindPattern.MatchString(row.kind) ||
		row.payloadRevision < 1 || !row.payloadReference.Valid ||
		len(row.payloadReference.String) < 1 ||
		len(row.payloadReference.String) > jobsv1.MaximumReferenceBytes ||
		!claimDigestPattern.MatchString(row.payloadSHA256) || !effectValid {
		return JobClaim{}, ErrJobClaimInvalid
	}
	return JobClaim{
		authority: authority, attemptID: row.attemptID, jobID: row.jobID,
		tenantID: tenantID, tenantScoped: tenantValid, kind: row.kind,
		payloadRevision: int(row.payloadRevision), payloadReference: row.payloadReference.String,
		payloadSHA256: row.payloadSHA256, effectClass: effectClass,
		effectKey: row.effectKey.String, hasEffectKey: row.effectKey.Valid,
		priority: int(row.priority), runAt: row.runAt, attempt: int(row.attempt),
		maximumAttempts: int(row.maximumAttempts), leaseOwnerID: row.leaseOwnerID,
		leaseGeneration: row.leaseGeneration, startedAt: row.startedAt,
		leaseUntil: row.leaseUntil,
	}, nil
}

func (claim JobClaim) Authority() JobAuthority         { return claim.authority }
func (claim JobClaim) AttemptID() uuid.UUID            { return claim.attemptID }
func (claim JobClaim) JobID() uuid.UUID                { return claim.jobID }
func (claim JobClaim) TenantID() (uuid.UUID, bool)     { return claim.tenantID, claim.tenantScoped }
func (claim JobClaim) Kind() string                    { return claim.kind }
func (claim JobClaim) PayloadRevision() int            { return claim.payloadRevision }
func (claim JobClaim) PayloadReference() string        { return claim.payloadReference }
func (claim JobClaim) PayloadSHA256() string           { return claim.payloadSHA256 }
func (claim JobClaim) EffectClass() jobsv1.EffectClass { return claim.effectClass }
func (claim JobClaim) EffectIdempotencyKey() (string, bool) {
	return claim.effectKey, claim.hasEffectKey
}
func (claim JobClaim) Priority() int           { return claim.priority }
func (claim JobClaim) RunAt() time.Time        { return claim.runAt }
func (claim JobClaim) Attempt() int            { return claim.attempt }
func (claim JobClaim) MaximumAttempts() int    { return claim.maximumAttempts }
func (claim JobClaim) LeaseOwnerID() uuid.UUID { return claim.leaseOwnerID }
func (claim JobClaim) LeaseGeneration() int64  { return claim.leaseGeneration }
func (claim JobClaim) StartedAt() time.Time    { return claim.startedAt }
func (claim JobClaim) LeaseUntil() time.Time   { return claim.leaseUntil }

func (claim JobClaim) valid(owner *workerIdentity) bool {
	return owner != nil && owner.marker == 1 && claim.authority.valid(owner) &&
		claim.authority.owner == owner && claim.attemptID != uuid.Nil &&
		claim.jobID != uuid.Nil && claim.leaseOwnerID != uuid.Nil &&
		claim.attempt >= 1 && claim.attempt <= claim.maximumAttempts &&
		claim.leaseGeneration == int64(claim.attempt) && !claim.startedAt.IsZero() &&
		claim.leaseUntil.Equal(claim.startedAt.Add(jobsv1.LeaseSeconds*time.Second))
}
