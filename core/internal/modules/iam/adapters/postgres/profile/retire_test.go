package profile

import (
	"context"
	"errors"
	"testing"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func TestRetireRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Retire(
		context.Background(), tenantScope{}, profilecommand.RetireCommand{},
	)
	if !errors.Is(err, profilecommand.ErrInvalidCommand) {
		t.Fatalf("invalid retirement error = %v", err)
	}
}
