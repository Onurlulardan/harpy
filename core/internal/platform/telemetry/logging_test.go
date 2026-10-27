package telemetry

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func TestLoggerWritesStructuredRedactedJSON(t *testing.T) {
	info, err := buildinfo.Read(runtimev1.ServiceCoreAPI, runtimev1.EnvironmentDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	logger, err := NewLogger(&output, config.LogDebug, info)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("configured",
		"database_dsn", "must-not-appear",
		"nested", slog.GroupValue(slog.String("access_token", "also-secret")),
	)
	if strings.Contains(output.String(), "must-not-appear") || strings.Contains(output.String(), "also-secret") {
		t.Fatalf("structured log leaked secret: %s", output.String())
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["service.name"] != "core-api" || record["database_dsn"] != "[REDACTED]" {
		t.Fatalf("unexpected structured record: %#v", record)
	}
}

func TestSecretImplementsSlogRedaction(t *testing.T) {
	info, err := buildinfo.Read(runtimev1.ServiceCoreAPI, runtimev1.EnvironmentDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	logger, err := NewLogger(&output, config.LogInfo, info)
	if err != nil {
		t.Fatal(err)
	}
	loaded := testAPIConfig(t)
	logger.Info("config", "value", loaded.LongRunning().Runtime().Postgres().DSN())
	if strings.Contains(output.String(), "example.invalid") || !strings.Contains(output.String(), "[REDACTED]") {
		t.Fatalf("secret LogValuer failed: %s", output.String())
	}
}

func testAPIConfig(t *testing.T) config.API {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", "postgresql://user@example.invalid/harpy?sslmode=disable")
	setTelemetryCursorKeyring(t)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}
