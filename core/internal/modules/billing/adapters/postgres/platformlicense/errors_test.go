package platformlicense

import (
	"errors"
	"testing"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
)

type stateError string

func (err stateError) Error() string    { return string(err) }
func (err stateError) SQLState() string { return string(err) }

func TestClassifyKeepsIdempotencyAndAuthorityConflictsDistinct(t *testing.T) {
	if err := classify(stateError("23505")); !errors.Is(
		err, platformlicenseapp.ErrIdempotencyReused,
	) {
		t.Fatalf("idempotency error=%v", err)
	}
	if err := classify(stateError("55000")); !errors.Is(
		err, platformlicenseapp.ErrConflict,
	) {
		t.Fatalf("authority conflict error=%v", err)
	}
}
