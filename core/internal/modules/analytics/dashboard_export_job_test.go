package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func TestDashboardExportPlannerRejectsDriftAndPinsEffectIdentity(t *testing.T) {
	t.Parallel()
	claim := dashboardExportClaimFixture()
	planner := dashboardExportJobPlanner{kind: DashboardExportDeleteJobKind,
		operation: "delete", namespace: dashboardExportDeleteEffectNamespace}
	spec, err := planner.plan(claim)
	if err != nil || spec.Provider() != dashboardExportEffectProvider ||
		spec.Operation() != "delete" {
		t.Fatalf("plan() = %#v, %v", spec, err)
	}
	if _, err = (dashboardExportJobPlanner{kind: DashboardExportRenderJobKind,
		operation: "render", namespace: uuid.New()}).plan(claim); err == nil {
		t.Fatal("generic planner accepted delegated render work")
	}
	claim.digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err = planner.plan(claim); err == nil {
		t.Fatal("planner accepted drifted export payload evidence")
	}
}

func TestDashboardExportDeleteProviderMapsDurableWorkerOutcomes(t *testing.T) {
	t.Parallel()
	worker := &dashboardExportJobWorkerFixture{result: exportapp.RenderResult{
		State: exportapp.StateDeleted}}
	provider, err := newDashboardExportJobProvider(DashboardExportDeleteJobKind,
		"delete", worker, func() time.Time {
			return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
		})
	if err != nil {
		t.Fatal(err)
	}
	request := dashboardExportEffectRequest(t, DashboardExportDeleteJobKind, "delete")
	result := provider.Execute(context.Background(), request)
	receipt, ok := result.Receipt()
	if !ok || receipt.Status != effectsv1.ReceiptSucceeded ||
		receipt.Code != "ANALYTICS_EXPORT_DELETED" || worker.calls != 1 {
		t.Fatalf("provider result = %#v/%v, calls=%d", receipt, ok, worker.calls)
	}
	worker.err = exportapp.ErrOutcomeUnknown
	if result = provider.Execute(context.Background(), request); !result.IsUnknown() {
		t.Fatal("provider converted an unknown object outcome into a retryable receipt")
	}
}

func TestDashboardExportGenericProviderRejectsRender(t *testing.T) {
	t.Parallel()
	worker := &dashboardExportJobWorkerFixture{}
	provider, err := newDashboardExportJobProvider(DashboardExportRenderJobKind,
		"render", worker, time.Now)
	if provider != nil || !errors.Is(err, exportapp.ErrUnavailable) {
		t.Fatalf("generic render provider = %#v, %v", provider, err)
	}
}

func TestDashboardExportExpiryProviderAcceptsCleanupTerminalStates(t *testing.T) {
	t.Parallel()
	for _, state := range []exportapp.State{exportapp.StateDeleted,
		exportapp.StateExpired, exportapp.StateCancelled, exportapp.StateFailed} {
		worker := &dashboardExportJobWorkerFixture{result: exportapp.RenderResult{State: state}}
		provider, err := newDashboardExportJobProvider(DashboardExportExpireJobKind,
			"expire", worker, func() time.Time {
				return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
			})
		if err != nil {
			t.Fatal(err)
		}
		result := provider.Execute(context.Background(),
			dashboardExportEffectRequest(t, DashboardExportExpireJobKind, "expire"))
		receipt, ok := result.Receipt()
		if !ok || receipt.Status != effectsv1.ReceiptSucceeded ||
			receipt.Code != "ANALYTICS_EXPORT_EXPIRY_COMPLETE" || worker.calls != 1 {
			t.Fatalf("expiry state %s result = %#v/%v calls=%d",
				state, receipt, ok, worker.calls)
		}
	}
}

type dashboardExportClaimStub struct {
	tenantID, attemptID, jobID uuid.UUID
	kind, reference, digest    string
}

func dashboardExportClaimFixture() dashboardExportClaimStub {
	reference := "019b1000-0000-7000-8000-000000000061"
	return dashboardExportClaimStub{
		tenantID:  uuid.MustParse("019b1000-0000-7000-8000-000000000001"),
		attemptID: uuid.MustParse("019b1000-0000-7000-8000-000000000071"),
		jobID:     uuid.MustParse("019b1000-0000-7000-8000-000000000062"),
		kind:      DashboardExportDeleteJobKind, reference: reference,
		digest: dashboardExportReferenceDigest(reference),
	}
}

func (claim dashboardExportClaimStub) AttemptID() uuid.UUID { return claim.attemptID }
func (claim dashboardExportClaimStub) JobID() uuid.UUID     { return claim.jobID }
func (claim dashboardExportClaimStub) TenantID() (uuid.UUID, bool) {
	return claim.tenantID, true
}
func (claim dashboardExportClaimStub) Kind() string              { return claim.kind }
func (dashboardExportClaimStub) PayloadRevision() int            { return 1 }
func (claim dashboardExportClaimStub) PayloadReference() string  { return claim.reference }
func (claim dashboardExportClaimStub) PayloadSHA256() string     { return claim.digest }
func (dashboardExportClaimStub) EffectClass() jobsv1.EffectClass { return jobsv1.EffectExternal }
func (claim dashboardExportClaimStub) EffectIdempotencyKey() (string, bool) {
	return claim.kind + ":" + claim.reference, true
}

type dashboardExportJobWorkerFixture struct {
	result exportapp.RenderResult
	err    error
	calls  int
}

func (worker *dashboardExportJobWorkerFixture) ProcessDelete(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID,
) (exportapp.RenderResult, error) {
	worker.calls++
	return worker.result, worker.err
}

func (worker *dashboardExportJobWorkerFixture) ProcessExpiry(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID,
) (exportapp.RenderResult, error) {
	worker.calls++
	return worker.result, worker.err
}

func dashboardExportEffectRequest(t *testing.T, kind, operation string) effectsv1.Request {
	t.Helper()
	tenant := "019b1000-0000-7000-8000-000000000001"
	exportID := "019b1000-0000-7000-8000-000000000061"
	jobID := "019b1000-0000-7000-8000-000000000062"
	authority := effectsv1.Authority{ScopeKind: effectsv1.ScopeTenant,
		ScopeID: tenant, GroupID: "analytics", GroupGeneration: 1,
		Domain: "analytics", CutoverGeneration: 1, WriterOrigin: effectsv1.WriterV2}
	request, err := effectsv1.NewRequest(
		uuid.NewSHA1(dashboardExportReceiptNamespace,
			[]byte(kind+"\x00"+operation)).String(), kind, 1, &tenant, authority,
		dashboardExportEffectProvider, operation, kind+":"+exportID, jobID, jobID,
		exportID, dashboardExportReferenceDigest(exportID),
		"2026-08-02T11:59:00Z")
	if err != nil {
		t.Fatal(err)
	}
	return request
}
