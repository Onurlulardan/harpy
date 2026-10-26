package billing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type freeBaseReader struct{ pin SaaSFreeBasePin }

func (reader freeBaseReader) FetchSaaSFreeBase(context.Context) (SaaSFreeBasePin, error) {
	return reader.pin, nil
}

func TestReadSaaSFreeBaseRequiresExactReleasePin(t *testing.T) {
	now := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	version := uuid.MustParse("019b0000-0000-7000-8000-000000000201")
	pin, err := ReadSaaSFreeBase(context.Background(), freeBaseReader{SaaSFreeBasePin{
		VersionID: version, Revision: 1, SHA256: strings.Repeat("a", 64), EffectiveAt: now,
	}})
	if err != nil || !pin.IsValid() || pin.VersionID != version {
		t.Fatalf("pin = %#v, %v", pin, err)
	}
	invalid := freeBaseReader{SaaSFreeBasePin{
		VersionID: version, Revision: 2, SHA256: strings.Repeat("a", 64), EffectiveAt: now,
	}}
	if _, err := ReadSaaSFreeBase(context.Background(), invalid); err == nil {
		t.Fatal("mutable free-base revision was accepted")
	}
}
