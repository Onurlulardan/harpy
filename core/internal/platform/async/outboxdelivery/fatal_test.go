package outboxdelivery

import (
	"errors"
	"fmt"
	"testing"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func TestFatalSessionErrorsSeparateContractFailureFromDependencyFailure(t *testing.T) {
	if !fatalSessionError(fmt.Errorf("wrapped: %w", asyncpostgres.ErrPublishClaimInvalid)) {
		t.Fatal("invalid claimed event was treated as reconnectable")
	}
	if fatalSessionError(errors.New("database temporarily unavailable")) {
		t.Fatal("transient dependency failure was treated as fatal")
	}
	if fatalSessionError(asyncpostgres.ErrOutboxAuthorityStale) {
		t.Fatal("expected cutover race was treated as fatal")
	}
}
