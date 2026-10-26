package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestRevokeACLFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).RevokeACL(context.Background(), app.ACLMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
