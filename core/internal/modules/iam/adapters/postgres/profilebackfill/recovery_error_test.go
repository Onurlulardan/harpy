package profilebackfill

import (
	"errors"
	"fmt"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

type recoveryStateFixture string

func (state recoveryStateFixture) Error() string    { return "database failure" }
func (state recoveryStateFixture) SQLState() string { return string(state) }

func TestRecoveryStoreErrorClassifiesWrappedNoForkConflict(t *testing.T) {
	for _, state := range []recoveryStateFixture{"P5742", "40001"} {
		wrapped := fmt.Errorf("wrapped: %w", state)
		if !errors.Is(recoveryStoreError(wrapped), profilebackfillapp.ErrConflict) {
			t.Fatalf("wrapped %s conflict escaped stable sentinel", state)
		}
	}
	unknown := recoveryStateFixture("55000")
	if !errors.Is(recoveryStoreError(unknown), profilebackfillapp.ErrUnavailable) ||
		!errors.Is(recoveryStoreError(unknown), unknown) {
		t.Fatal("unknown database cause was not retained")
	}
}
