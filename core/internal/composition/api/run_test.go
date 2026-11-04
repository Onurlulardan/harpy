package api

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func TestRunReachesDatabaseWithoutSchemaContract(t *testing.T) {
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("DEPLOYMENT_MODE", "saas")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_v2_runtime@127.0.0.1:1/harpy?sslmode=disable",
	)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", apiTestCursorKeyring)
	configfixture.SetAPI(t)
	var stderr bytes.Buffer
	code := run(context.Background(), nil, &stderr)
	if code != processcode.DependencyPreflight {
		t.Fatalf("run code=%d stderr=%s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "schema contract") {
		t.Fatalf("schema contract is still a startup gate: %s", stderr.String())
	}
}
