package rehydraterequestv1

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestParseBindsExactKitAndDisposition(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 123456000, time.UTC)
	kit := []byte("exact-kit")
	request := Request{
		RecoverySetID:         "018f0000-0000-7000-8000-000000000010",
		ExpectedEnvironmentID: "018f0000-0000-7000-8000-000000000001",
		TargetDisposition:     "empty",
		WaiverReason:          "empty_no_database_or_policy",
		ProbeSHA256:           hex.EncodeToString(make([]byte, 32)),
		RequestedAt:           now.Format(time.RFC3339Nano),
	}
	document, err := Build(request, kit, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(document, kit, now); err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(document, []byte("other"), now); err == nil {
		t.Fatal("different recovery kit was accepted")
	}
}
