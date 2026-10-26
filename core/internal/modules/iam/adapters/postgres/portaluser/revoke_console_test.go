package portaluserpostgres

import (
	"context"
	"errors"
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func TestRevokeRejectsUnboundScope(t *testing.T) {
	_, err := (&store{}).RevokeConsole(context.Background(), tenantScope{}, portaluser.RevisionCommand{})
	if !errors.Is(err, portaluser.ErrInvalid) {
		t.Fatalf("RevokeConsole error = %v, want ErrInvalid", err)
	}
}
