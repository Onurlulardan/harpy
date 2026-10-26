package postgres

import (
	"context"
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func TestIAMCredentialEventWriterRejectsActorAndTenantSubstitution(t *testing.T) {
	subjectID := uuid.MustParse("ac000000-0000-4000-8000-000000000001")
	writer := &iamCredentialEventWriter{
		writer: &repository{authority: authorityStamp{
			scopeID: uuid.MustParse(consumerTenantA), groupID: "iam_system_resources",
			groupGeneration: 1, domain: "iam", generation: 1,
		}},
		subjectID: subjectID,
	}
	if _, err := writer.EnqueueEvent(context.Background(), eventsv1.Envelope{
		ActorID: uuid.NewString(),
	}); !errors.Is(err, ErrAuthorityMismatch) {
		t.Fatalf("actor substitution error = %v", err)
	}
	if _, err := writer.EnqueueEvent(context.Background(), eventsv1.Envelope{
		ActorID: subjectID.String(), TenantID: pointerString(uuid.NewString()),
	}); !errors.Is(err, ErrAuthorityMismatch) {
		t.Fatalf("tenant substitution error = %v", err)
	}
}
