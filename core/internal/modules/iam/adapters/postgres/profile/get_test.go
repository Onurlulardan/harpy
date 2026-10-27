package profile

import (
	"context"
	"errors"
	"testing"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func TestGetSelfRejectsUnboundScopeBeforeQuery(t *testing.T) {
	_, err := (*commandStore)(nil).GetSelf(
		context.Background(), tenantScope{}, profilecommand.SelfEvidence{
			ActorID:              uuid.MustParse("51000000-0000-4000-8000-000000000001"),
			ExpectedTokenVersion: 2, ExpectedAuthorizationVersion: 3,
		},
	)
	if !errors.Is(err, profilecommand.ErrInvalidCommand) {
		t.Fatalf("unbound self read error = %v", err)
	}
}
