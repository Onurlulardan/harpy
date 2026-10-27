package postgres

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestClaimJobsRejectsForgedAuthorityAndInvalidBatchBeforeTransaction(t *testing.T) {
	t.Parallel()
	database := &workerTestDatabase{}
	repository := newWorkerTestRepository()
	repository.database = database
	authority := validWorkerAuthority(t, repository)
	if _, err := repository.ClaimJobs(
		context.Background(), JobAuthority{}, uuid.New(), 1, validClaimCatalog(t),
	); !errors.Is(err, ErrJobAuthorityInvalid) {
		t.Fatalf("forged authority error = %v", err)
	}
	for _, limit := range []int{0, 51} {
		if _, err := repository.ClaimJobs(
			context.Background(), authority, uuid.New(), limit, validClaimCatalog(t),
		); !errors.Is(err, ErrJobClaimLimit) {
			t.Fatalf("claim limit %d error = %v", limit, err)
		}
	}
	if database.beginCalls != 0 {
		t.Fatalf("invalid claim began %d transactions", database.beginCalls)
	}
	for _, limit := range []int{1, 50} {
		_, err := repository.ClaimJobs(
			context.Background(), authority, uuid.New(), limit, validClaimCatalog(t),
		)
		if !errors.Is(err, errWorkerTestDatabase) {
			t.Fatalf("claim boundary %d error = %v", limit, err)
		}
	}
	if database.beginCalls != 2 {
		t.Fatalf("valid claim boundaries began %d transactions", database.beginCalls)
	}
}

func TestClaimJobsRollsBackScanAndCommitErrors(t *testing.T) {
	t.Parallel()
	scanFailure := errors.New("scan failed")
	commitFailure := errors.New("commit failed")
	tests := []struct {
		name string
		tx   *workerTestTransaction
		want error
	}{
		{
			name: "scan",
			tx: &workerTestTransaction{queryRows: &workerTestRows{
				remaining: 1, scanErr: scanFailure,
			}},
			want: scanFailure,
		},
		{
			name: "commit",
			tx: &workerTestTransaction{
				queryRows: &workerTestRows{}, commitErr: commitFailure,
			},
			want: commitFailure,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := newWorkerTestRepository()
			repository.database = &workerTestDatabase{transaction: test.tx}
			_, err := repository.ClaimJobs(
				context.Background(), validWorkerAuthority(t, repository), uuid.New(), 1,
				validClaimCatalog(t),
			)
			if !errors.Is(err, test.want) || test.tx.rollbackCalls != 1 {
				t.Fatalf("claim error = %v, rollbacks = %d", err, test.tx.rollbackCalls)
			}
			if len(test.tx.calls) < 3 || !slices.Equal(
				test.tx.calls[:2], []string{"assert authority", "lock jobs"},
			) || test.tx.calls[len(test.tx.calls)-1] != "rollback" {
				t.Fatalf("claim transaction order = %v", test.tx.calls)
			}
		})
	}
}

func TestClaimJobsReturnsOnlyAfterSuccessfulCommit(t *testing.T) {
	t.Parallel()
	transaction := &workerTestTransaction{queryRows: &workerTestRows{}}
	repository := newWorkerTestRepository()
	repository.database = &workerTestDatabase{transaction: transaction}
	claims, err := repository.ClaimJobs(
		context.Background(), validWorkerAuthority(t, repository), uuid.New(), 50,
		validClaimCatalog(t),
	)
	if err != nil || len(claims) != 0 || transaction.rollbackCalls != 0 ||
		!slices.Equal(transaction.calls, []string{"assert authority", "lock jobs", "commit"}) {
		t.Fatalf("successful claim = %v, %v, calls %v", claims, err, transaction.calls)
	}
}

func TestJobClaimIsImmutableAndLeaseFenced(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	authority := validWorkerAuthority(t, repository)
	workerID := uuid.New()
	startedAt := time.Now().UTC()
	row := validJobClaimRow(authority, workerID, startedAt)
	claim, err := newJobClaim(row, authority, workerID)
	if err != nil {
		t.Fatal(err)
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped || tenantID != authority.scopeID || claim.Attempt() != 1 ||
		claim.LeaseGeneration() != 1 || claim.EffectClass() != jobsv1.EffectDatabaseTransactional {
		t.Fatalf("claim projection = %#v", claim)
	}
	row.leaseUntil = startedAt.Add(59 * time.Second)
	if _, err := newJobClaim(row, authority, workerID); !errors.Is(err, ErrJobClaimInvalid) {
		t.Fatalf("invalid lease error = %v", err)
	}
}

func TestClassifyJobAuthorityErrorIsTyped(t *testing.T) {
	t.Parallel()
	for _, code := range []string{"P5601", "P5602", "P5603"} {
		err := classifyJobAuthorityError(&pgconn.PgError{Code: code})
		if !errors.Is(err, ErrJobAuthorityStale) {
			t.Fatalf("SQLSTATE %s error = %v", code, err)
		}
	}
}

func validWorkerAuthority(t *testing.T, repository *workerRepository) JobAuthority {
	t.Helper()
	authority, err := newJobAuthority(
		repository.identity, "tenant", uuid.New(), "forms", 3, "forms", 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	return authority
}

func validJobClaimRow(
	authority JobAuthority,
	workerID uuid.UUID,
	startedAt time.Time,
) jobClaimRow {
	return jobClaimRow{
		attemptID: uuid.Must(uuid.NewV7()), jobID: uuid.New(),
		tenantID:  pgtype.UUID{Bytes: [16]byte(authority.scopeID), Valid: true},
		scopeKind: authority.scopeKind, scopeID: authority.scopeID,
		groupID: authority.groupID, groupGeneration: authority.groupGeneration,
		domain: authority.domain, generation: authority.generation,
		kind: "forms.export", payloadRevision: 1,
		payloadReference: pgtype.Text{String: "jobs/payload/1", Valid: true},
		payloadSHA256:    strings.Repeat("a", 64),
		effectClass:      string(jobsv1.EffectDatabaseTransactional), priority: 10,
		runAt: startedAt.Add(-time.Second), attempt: 1, maximumAttempts: 8,
		leaseOwnerID: workerID, leaseGeneration: 1, startedAt: startedAt,
		leaseUntil: startedAt.Add(60 * time.Second),
	}
}
