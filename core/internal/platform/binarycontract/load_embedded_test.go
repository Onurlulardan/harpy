package binarycontract

import (
	"errors"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestLoadEmbeddedPropagatesMissingContract(t *testing.T) {
	artifact, err := loadEmbedded(func() ([]byte, error) {
		return nil, runtimev1.BinarySchemaContractNotEmbeddedError{}
	})
	var missing runtimev1.BinarySchemaContractNotEmbeddedError
	if !errors.As(err, &missing) {
		t.Fatalf("expected missing contract error, got %v", err)
	}
	if artifact.IsVerified() {
		t.Fatal("missing embedded contract returned verified authority")
	}
}
