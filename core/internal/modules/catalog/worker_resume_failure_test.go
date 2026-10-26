package catalog

import (
	"errors"
	"testing"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func TestSettleLostDoesNotKillWorkerOnUnknownResumeErrors(t *testing.T) {
	if err := settleLost(errors.New("catalog resume exploded")); err != nil {
		t.Fatalf("unknown resume error killed worker: %v", err)
	}
	if err := settleLost(asyncpostgres.ErrJobLeaseLost); !errors.Is(err, asyncpostgres.ErrJobLeaseLost) {
		t.Fatalf("lease loss was swallowed: %v", err)
	}
	if err := settleLost(nil); err != nil {
		t.Fatal(err)
	}
}
