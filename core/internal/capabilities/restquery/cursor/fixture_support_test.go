package cursor

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func requireQueryCode(t testing.TB, err error, code restqueryv1.ErrorCode) {
	t.Helper()
	queryError, ok := err.(api.QueryError)
	if !ok || queryError.Code() != code {
		t.Fatalf("error = %T %v, want %s", err, err, code)
	}
}
