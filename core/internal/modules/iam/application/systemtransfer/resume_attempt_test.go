package systemtransfer

import (
	"context"
	"crypto/sha256"
	"io"
	"strings"
	"testing"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type resumeLoader struct{ state AttemptState }

func (loader *resumeLoader) LoadAttempt(
	context.Context,
	AttemptClaim,
) (AttemptState, error) {
	return loader.state, nil
}

type resumeSource struct {
	document string
	opened   ExactImportSource
	calls    int
}

func (source *resumeSource) OpenExactImportSource(
	_ context.Context,
	exact ExactImportSource,
) (io.ReadCloser, error) {
	source.calls++
	source.opened = exact
	return io.NopCloser(strings.NewReader(source.document)), nil
}

func TestResumeAttemptContinuesAfterDurableRowWithoutReplayingMutation(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Organizations)
	state := resumeAttemptState(attempt, 1, false)
	loader := &resumeLoader{state: state}
	source := &resumeSource{document: strings.Join([]string{
		`{"name":"Operations","slug":"operations"}`,
		`{"name":"Finance","slug":"finance"}`,
		"",
	}, "\n")}
	result, err := fixtureMutationResult(
		publicresource.Organizations, publicresource.ID(fixtureRecordID), 1, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	resources := &fixtureResourcePort{result: result}
	receipts := &fixtureReceiptStore{}
	worker, err := NewWorker(&fixtureScopes{}, &fixtureAuthorizer{
		tenantID: fixtureTenantID, actorID: fixtureActorID, version: 7,
	}, &fixtureIdentityLookup{}, resources, &fixtureInvitation{}, receipts)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewSystemImportRuntime(loader, source, worker)
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := runtime.ResumeAttempt(t.Context(), resumeClaim(attempt.Resource))
	if err != nil || resumed.ID != attempt.ID || resources.createCalls != 1 ||
		len(receipts.appended) != 1 || receipts.appended[0].RowIndex != 2 ||
		source.calls != 1 || source.opened.SourceID != state.Source.SourceID {
		t.Fatalf("resumed=%+v creates=%d receipts=%+v source=%d err=%v",
			resumed, resources.createCalls, receipts.appended, source.calls, err)
	}
}

func TestResumeAttemptAcceptsAtomicTerminalReplayWithoutOpeningArtifact(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Roles)
	loader := &resumeLoader{state: resumeAttemptState(attempt, 2, true)}
	source := &resumeSource{document: "not-opened"}
	worker, _ := NewWorker(&fixtureScopes{}, &fixtureAuthorizer{},
		&fixtureIdentityLookup{}, &fixtureResourcePort{}, &fixtureInvitation{},
		&fixtureReceiptStore{})
	runtime, _ := NewSystemImportRuntime(loader, source, worker)
	resumed, err := runtime.ResumeAttempt(t.Context(), resumeClaim(attempt.Resource))
	if err != nil || resumed.ID != attempt.ID || source.calls != 0 {
		t.Fatalf("resumed=%+v source calls=%d err=%v", resumed, source.calls, err)
	}
}

func resumeClaim(resource publicresource.Resource) AttemptClaim {
	return AttemptClaim{TenantID: fixtureTenantID, ImportID: fixtureTransferID,
		JobID: fixtureJobID, AttemptID: fixtureAttemptID,
		LeaseOwnerID: fixtureLeaseID, LeaseGeneration: 1,
		Resource: resource}
}

func resumeAttemptState(attempt Attempt, processed int64, finished bool) AttemptState {
	sourceID := fixtureTransferID
	one := sha256.Sum256([]byte("one"))
	two := sha256.Sum256([]byte("two"))
	three := sha256.Sum256([]byte("three"))
	return AttemptState{Attempt: attempt, NextRowIndex: processed + 1,
		Processed: processed, Finished: finished, Source: ExactImportSource{
			TenantID: attempt.TenantID, ActorID: attempt.ActorID, SourceID: sourceID,
			Resource: attempt.Resource, Format: "jsonl", Artifact: SourceArtifact{
				ObjectKey: "iam-import/one", ObjectVersion: "version-1", KeyID: "kek-1",
				ProviderRequestID: "provider-1", HeaderSHA256: one,
				CiphertextSHA256: two, PlaintextSHA256: three,
				CiphertextBytes: 80, PlaintextBytes: 79,
				WrappedDEK: make([]byte, 32), WrapNonce: [12]byte{1},
				UploadedAt: time.Date(2026, 7, 18, 1, 0, 0, 0, time.UTC),
			},
		}}
}
