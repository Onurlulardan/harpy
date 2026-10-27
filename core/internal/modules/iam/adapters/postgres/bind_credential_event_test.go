package postgres

import (
	"context"
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type credentialEventSinkFixture struct{}

func (credentialEventSinkFixture) Enqueue(context.Context, eventsv1.Envelope) error {
	return nil
}

func (credentialEventSinkFixture) EnqueueJob(
	context.Context,
	jobsv1.Enqueue,
) (string, error) {
	return "", nil
}

func TestBindIAMCredentialEventRejectsClosedScope(t *testing.T) {
	if _, err := BindCredentialEvent(
		platformtx.CredentialTenantScope{}, uuid.New(),
		func(any, CredentialEventAuthority) (CredentialEventSink, error) {
			return credentialEventSinkFixture{}, nil
		},
	); !errors.Is(err, platformtx.ErrScopeClosed) {
		t.Fatalf("closed scope error = %v", err)
	}
}
