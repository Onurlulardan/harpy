package migrator

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func TestRunRejectsInvalidInvocationBeforeConfiguration(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(context.Background(), nil, &stdout, &stderr)
	if code != processcode.InvalidStartup {
		t.Fatalf("code=%d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `"error_class":"invocation"`) {
		t.Fatalf("stable invocation failure missing: %s", stderr.String())
	}
}

func TestSchemaExecutorFailsBeforeDatabaseWithoutBuildContract(t *testing.T) {
	setSchemaExecutorTestEnvironment(t)
	called := 0
	loader := func() (schemacontract.EmbeddedContract, error) {
		called++
		return schemacontract.EmbeddedContract{}, errors.New("missing")
	}
	var stderr bytes.Buffer
	code := runSchemaExecutorWithContractLoader(
		context.Background(), &stderr, loader,
	)
	if code != processcode.InvalidStartup || called != 1 {
		t.Fatalf("code=%d loader calls=%d", code, called)
	}
	if !strings.Contains(stderr.String(), `"error_class":"build_contract"`) {
		t.Fatalf("stable build failure missing: %s", stderr.String())
	}
}

func setSchemaExecutorTestEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_v2_schema_executor@127.0.0.1:1/harpy?sslmode=disable",
	)
}
