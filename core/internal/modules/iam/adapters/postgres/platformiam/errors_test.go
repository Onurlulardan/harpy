package platformiampostgres

import (
	"errors"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

type stateError string

func (err stateError) Error() string    { return "database failure" }
func (err stateError) SQLState() string { return string(err) }

func TestStoreErrorMapsAuthorityAndCASFailures(t *testing.T) {
	for state, expected := range map[string]error{
		"42501": platformdomain.ErrDenied,
		"40001": platformdomain.ErrConflict,
		"23505": platformdomain.ErrConflict,
		"P1404": platformdomain.ErrNotFound,
		"XX000": platformdomain.ErrStorage,
	} {
		if actual := storeError(stateError(state)); !errors.Is(actual, expected) {
			t.Fatalf("state %s mapped to %v", state, actual)
		}
	}
}
