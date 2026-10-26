package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

type invitationSQLStateFixture string

func (state invitationSQLStateFixture) Error() string    { return "database failure" }
func (state invitationSQLStateFixture) SQLState() string { return string(state) }

func TestMapInvitationErrorClassifiesWrappedRecoveryFence(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", invitationSQLStateFixture("P5741"))
	if !errors.Is(mapInvitationError(err), invitation.ErrRecoveryRequired) {
		t.Fatalf("mapInvitationError() = %v", mapInvitationError(err))
	}
}
