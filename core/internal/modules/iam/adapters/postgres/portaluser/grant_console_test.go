package portaluserpostgres

import (
	"context"
	"errors"
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func TestGrantRejectsUnboundScope(t *testing.T) {
	_, err := (&store{}).GrantConsole(context.Background(), tenantScope{}, portaluser.RevisionCommand{})
	if !errors.Is(err, portaluser.ErrInvalid) {
		t.Fatalf("GrantConsole error = %v, want ErrInvalid", err)
	}
}
