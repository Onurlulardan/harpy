package adoption

import (
	"context"
	"errors"
	"testing"
)

func TestFinalizeRunRejectsMissingStore(t *testing.T) {
	_, err := FinalizeRun(context.Background(), nil, Run{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
