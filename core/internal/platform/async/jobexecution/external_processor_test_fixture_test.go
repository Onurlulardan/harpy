package jobexecution

import (
	"context"
	"errors"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

const (
	testTenantID  = "11111111-1111-4111-8111-111111111111"
	testEffectID  = "22222222-2222-4222-8222-222222222222"
	testCausation = "33333333-3333-4333-8333-333333333333"
	testReceiptID = "44444444-4444-4444-8444-444444444444"
	testDigest    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

type externalPlannerStub struct {
	spec  asyncpostgres.ExternalEffectRequestSpec
	err   error
	calls *[]string
}

func (stub *externalPlannerStub) Plan(
	context.Context,
	asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	appendCall(stub.calls, "plan")
	return stub.spec, stub.err
}

type preEffectStub struct {
	started  StartedEffect
	startErr error
	failErr  error
	calls    *[]string
}

func (stub *preEffectStub) Start(
	context.Context,
	asyncpostgres.JobClaim,
	asyncpostgres.ExternalEffectRequestSpec,
) (StartedEffect, error) {
	appendCall(stub.calls, "start")
	return stub.started, stub.startErr
}

func (stub *preEffectStub) FailBeforeEffect(
	ctx context.Context,
	_ asyncpostgres.JobClaim,
	_ asyncpostgres.JobFailure,
) (asyncpostgres.JobFailureResult, error) {
	if ctx == nil || ctx.Err() != nil {
		return asyncpostgres.JobFailureResult{}, errors.New("cancelled finalization context")
	}
	appendCall(stub.calls, "fail-before")
	return asyncpostgres.JobFailureResult{}, stub.failErr
}

type externalFinalizerStub struct {
	calls   *[]string
	err     error
	receipt effectsv1.Receipt
}

func (stub *externalFinalizerStub) CompleteExternalJob(
	ctx context.Context,
	_ asyncpostgres.ExternalEffectPermit,
	receipt effectsv1.Receipt,
) (time.Time, error) {
	if ctx.Err() != nil {
		return time.Time{}, ctx.Err()
	}
	appendCall(stub.calls, "complete")
	stub.receipt = receipt
	return time.Now(), stub.err
}

func (stub *externalFinalizerStub) FailExternalJob(
	ctx context.Context,
	_ asyncpostgres.ExternalEffectPermit,
	receipt effectsv1.Receipt,
	_ asyncpostgres.JobFailure,
) (asyncpostgres.JobFailureResult, error) {
	if ctx.Err() != nil {
		return asyncpostgres.JobFailureResult{}, ctx.Err()
	}
	appendCall(stub.calls, "fail")
	stub.receipt = receipt
	return asyncpostgres.JobFailureResult{}, stub.err
}

func (stub *externalFinalizerStub) MarkExternalJobOutcomeUnknown(
	ctx context.Context,
	_ asyncpostgres.ExternalEffectPermit,
) (time.Time, error) {
	if ctx.Err() != nil {
		return time.Time{}, ctx.Err()
	}
	appendCall(stub.calls, "unknown")
	return time.Now(), stub.err
}

type recordingJitter struct {
	bound uint64
	value uint64
}

func (jitter *recordingJitter) Uint64N(bound uint64) uint64 {
	jitter.bound = bound
	return jitter.value
}

func appendCall(calls *[]string, value string) {
	if calls != nil {
		*calls = append(*calls, value)
	}
}

func externalProcessorFixture(
	t *testing.T,
	_ *[]string,
) (jobsv1.KindPolicy, asyncpostgres.ExternalEffectRequestSpec, StartedEffect) {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(
		"notification.email", 1, "notifications", jobsv1.EffectExternal, 3,
	)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := asyncpostgres.NewExternalEffectRequestSpec(
		uuid.MustParse(testEffectID), "test.mail", "send", "corr-1",
		uuid.MustParse(testCausation),
	)
	if err != nil {
		t.Fatal(err)
	}
	request := externalRequestFixture(t)
	return policy, spec, StartedEffect{request: request}
}

func externalRequestFixture(t *testing.T) effectsv1.Request {
	t.Helper()
	tenant := testTenantID
	request, err := effectsv1.NewRequest(
		testEffectID, "notification.email", 1, &tenant,
		effectsv1.Authority{
			ScopeKind: effectsv1.ScopeTenant, ScopeID: testTenantID,
			GroupID: "notifications", GroupGeneration: 1,
			Domain: "notifications", CutoverGeneration: 1,
			WriterOrigin: effectsv1.WriterV2,
		}, "test.mail", "send", "mail:1", "corr-1", testCausation,
		"payload://mail/1", testDigest, "2026-07-15T10:00:00Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func definitiveResult(t *testing.T, receipt effectsv1.Receipt) effectexecution.Result {
	t.Helper()
	result, err := effectexecution.Definitive(receipt)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
