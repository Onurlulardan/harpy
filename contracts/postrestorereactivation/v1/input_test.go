package postrestorereactivationv1

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestAwaitInputIsCanonicalAndExactlyBound(t *testing.T) {
	environmentID, operationID := uuid.NewString(), uuid.NewString()
	document, err := BuildAwait(environmentID, operationID)
	if err != nil {
		t.Fatal(err)
	}
	value, err := ParseAwait(document)
	if err != nil || value.EnvironmentID != environmentID ||
		value.RestoreOperationID != operationID {
		t.Fatalf("await input = %#v, %v", value, err)
	}
	tampered := bytes.Replace(
		document, []byte(operationID), []byte(uuid.NewString()), 1,
	)
	if _, err = ParseAwait(append(tampered, '\n')); err == nil {
		t.Fatal("non-canonical await input accepted")
	}
}
