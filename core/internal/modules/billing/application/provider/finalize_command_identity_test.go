package providerapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFinalizeCommandRejectsGeneratedIdentityCollisionsBeforeStorage(t *testing.T) {
	request := FinalizeCommandRequest{
		TenantID:        uuid.New(),
		JobID:           uuid.MustParse("018f1000-0000-7000-8000-000000000091"),
		AttemptID:       uuid.MustParse("018f1000-0000-7000-8000-000000000092"),
		CommandID:       uuid.MustParse("018f1000-0000-7000-8000-000000000093"),
		LeaseGeneration: 1, PayloadRevision: 1,
		PayloadReference: "billing-provider-finalize/" +
			"018f1000-0000-7000-8000-000000000093",
		PayloadSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	uniqueOne := uuid.MustParse("018f1000-0000-7000-8000-000000000094")
	uniqueTwo := uuid.MustParse("018f1000-0000-7000-8000-000000000095")
	tests := []struct {
		name          string
		first, second uuid.UUID
	}{
		{"next_job", request.JobID, uniqueTwo},
		{"next_attempt", request.AttemptID, uniqueTwo},
		{"next_command", request.CommandID, uniqueTwo},
		{"reconcile_job", uniqueOne, request.JobID},
		{"reconcile_attempt", uniqueOne, request.AttemptID},
		{"reconcile_command", uniqueOne, request.CommandID},
		{"successor_reconcile_same", uniqueOne, uniqueOne},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &finalizerStoreStub{}
			values := []uuid.UUID{test.first, test.second}
			finalizer, err := NewCommandFinalizer(
				store,
				func() (uuid.UUID, error) {
					value := values[0]
					values = values[1:]
					return value, nil
				},
				func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) },
			)
			if err != nil {
				t.Fatal(err)
			}
			_, err = finalizer.FinalizeCommand(context.Background(), request)
			if !errors.Is(err, ErrConfiguration) || store.called {
				t.Fatalf("FinalizeCommand() error=%v store called=%v", err, store.called)
			}
		})
	}
}
