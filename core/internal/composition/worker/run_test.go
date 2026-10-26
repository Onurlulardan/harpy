package worker

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func TestRunFailsBeforeDatabaseWhenBuildContractIsUnavailable(t *testing.T) {
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_v2_runtime@127.0.0.1:1/harpy?sslmode=disable",
	)
	t.Setenv("HARPY_RABBITMQ_URL", "amqp://worker:secret@127.0.0.1:1/%2f")
	configfixture.SetWorker(t)
	called := 0
	loader := func() (schemacontract.EmbeddedContract, error) {
		called++
		return schemacontract.EmbeddedContract{}, errors.New("missing")
	}
	var stderr bytes.Buffer
	code := runWithContractLoader(context.Background(), nil, &stderr, loader)
	if code != processcode.InvalidStartup || called != 1 {
		t.Fatalf("run code=%d loader calls=%d", code, called)
	}
	if !strings.Contains(stderr.String(), `"error_class":"build_contract"`) {
		t.Fatalf("stable build failure missing: %s", stderr.String())
	}
}
