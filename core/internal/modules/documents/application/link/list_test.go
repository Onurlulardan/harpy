package link

import (
	"context"
	"errors"
	"testing"
)

func TestListOperationContract(t *testing.T) {
	if any((*Application).ListReferenceBlockers) == nil {
		t.Fatal("reference blocker list operation is unavailable")
	}
}

func TestListFormRecordRejectsMissingStore(t *testing.T) {
	if _, err := listFormRecordLinks(
		context.Background(), nil, Source{},
	); !errors.Is(err, ErrInvalid) {
		t.Fatalf("list error = %v", err)
	}
}
