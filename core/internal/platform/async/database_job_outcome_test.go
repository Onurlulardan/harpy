package async

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDatabaseJobOutcomeSealsCanonicalSuccessor(t *testing.T) {
	runAt := time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	input := DatabaseJobSuccessorInput{
		JobID: uuid.MustParse("019c6a7e-7c00-7000-8000-000000000001"),
		RunAt: runAt, PayloadReference: "urn:harpy:iam:test",
		PayloadSHA256: strings.Repeat("a", 64), DedupeKey: "iam:test:1",
	}
	outcome, err := ContinueDatabaseJob(input)
	if err != nil || !outcome.Valid() {
		t.Fatalf("outcome = %+v, %v", outcome, err)
	}
	got, ok := outcome.Successor()
	if !ok || got != input {
		t.Fatalf("successor = %+v, %v", got, ok)
	}
	if !CompleteDatabaseJob().Valid() {
		t.Fatal("terminal database outcome is invalid")
	}
}

func TestDatabaseJobOutcomeRejectsAmbiguousOrUnboundedInput(t *testing.T) {
	valid := DatabaseJobSuccessorInput{
		JobID:            uuid.MustParse("019c6a7e-7c00-7000-8000-000000000001"),
		RunAt:            time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
		PayloadReference: "urn:harpy:iam:test",
		PayloadSHA256:    strings.Repeat("a", 64), DedupeKey: "iam:test:1",
	}
	tests := map[string]func(*DatabaseJobSuccessorInput){
		"not v7": func(value *DatabaseJobSuccessorInput) {
			value.JobID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
		},
		"local time": func(value *DatabaseJobSuccessorInput) {
			value.RunAt = value.RunAt.In(time.FixedZone("other", 3600))
		},
		"uppercase digest": func(value *DatabaseJobSuccessorInput) {
			value.PayloadSHA256 = strings.Repeat("A", 64)
		},
		"padded reference": func(value *DatabaseJobSuccessorInput) {
			value.PayloadReference = " padded"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			input := valid
			mutate(&input)
			if _, err := ContinueDatabaseJob(input); err == nil {
				t.Fatal("invalid successor was accepted")
			}
		})
	}
}
