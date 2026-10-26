package operatorv1

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestInitializeResultIsExactAndNonSecret(t *testing.T) {
	result, err := NewInitializeResult(InitializeCompleted,
		uuid.MustParse("019b3a20-0000-7000-8000-000000000001"),
		uuid.MustParse("019b3a20-0000-7000-8000-000000000002"))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := WriteInitializeResult(&output, result); err != nil {
		t.Fatal(err)
	}
	want := `{"contract_version":"operator-initialize-result.v1","status":"completed",` +
		`"operation_id":"019b3a20-0000-7000-8000-000000000001",` +
		`"tenant_id":"019b3a20-0000-7000-8000-000000000002"}` + "\n"
	if output.String() != want || strings.Contains(output.String(), "password") {
		t.Fatalf("result = %q", output.String())
	}
}

func TestInitializeResultRejectsInvalidState(t *testing.T) {
	if _, err := NewInitializeResult("failed", uuid.New(), uuid.New()); err == nil {
		t.Fatal("invalid status was accepted")
	}
}
