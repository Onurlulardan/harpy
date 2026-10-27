package profilepostgres

import (
	"errors"
	"fmt"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type profileStateFixture string

func (state profileStateFixture) Error() string    { return "database failure" }
func (state profileStateFixture) SQLState() string { return string(state) }

func TestProfileContractErrorClassifiesRecoveryWithoutMessageParsing(t *testing.T) {
	classified := profileContractError(profileStateFixture("P5741"))
	if !errors.Is(classified, profileapp.ErrRecoveryRequired) {
		t.Fatalf("profileContractError() = %v", classified)
	}
	if !errors.Is(profileContractError(profileStateFixture("55000")),
		identitydomain.ErrStorage) {
		t.Fatal("generic database failure escaped storage sentinel")
	}
	wrapped := fmt.Errorf("wrapped: %w", profileStateFixture("P5741"))
	if !errors.Is(profileContractError(wrapped), profileapp.ErrRecoveryRequired) {
		t.Fatal("wrapped recovery fence escaped stable sentinel")
	}
}
