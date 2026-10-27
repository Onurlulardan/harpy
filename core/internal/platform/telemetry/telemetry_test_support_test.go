package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func telemetryBuildInfo(t *testing.T) buildinfo.Info {
	t.Helper()
	info, err := buildinfo.Read(runtimev1.ServiceCoreAPI, runtimev1.EnvironmentTest)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func telemetrySettings(t *testing.T, enabled bool) config.Telemetry {
	t.Helper()
	if enabled {
		return telemetrySettingsForEndpoint(t, "http://collector.invalid:4318", "")
	}
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", "postgresql://user@example.invalid/harpy?sslmode=disable")
	setTelemetryCursorKeyring(t)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	return loaded.LongRunning().Runtime().Telemetry()
}

func telemetrySettingsForEndpoint(t *testing.T, endpoint, headers string) config.Telemetry {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", "postgresql://user@example.invalid/harpy?sslmode=disable")
	setTelemetryCursorKeyring(t)
	t.Setenv("HARPY_OTLP_ENDPOINT", endpoint)
	t.Setenv("HARPY_OTLP_INSECURE", "true")
	t.Setenv("HARPY_TRACE_SAMPLE_RATIO", "1")
	if headers != "" {
		t.Setenv("HARPY_OTLP_HEADERS", headers)
	}
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	return loaded.LongRunning().Runtime().Telemetry()
}

func setTelemetryCursorKeyring(t *testing.T) {
	t.Helper()
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", telemetryTestCursorKeyring)
}

const telemetryTestCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

func testLogger(output *lockedBuffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, nil))
}

type lockedBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (buffer *lockedBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	buffer.data = append(buffer.data, data...)
	return len(data), nil
}

func (buffer *lockedBuffer) String() string {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return string(buffer.data)
}

type fakeSpanExporter struct {
	mu          sync.Mutex
	exportErr   error
	shutdownErr error
	spans       []sdktrace.ReadOnlySpan
	shutdowns   int
}

func (exporter *fakeSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	exporter.spans = append(exporter.spans, spans...)
	return exporter.exportErr
}

func (exporter *fakeSpanExporter) Shutdown(context.Context) error {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	exporter.shutdowns++
	return exporter.shutdownErr
}

func (exporter *fakeSpanExporter) snapshot() (int, int) {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	return len(exporter.spans), exporter.shutdowns
}

type fakeMetricExporter struct {
	mu          sync.Mutex
	exportErr   error
	flushErr    error
	shutdownErr error
	exports     int
	flushes     int
	shutdowns   int
}

func (*fakeMetricExporter) Temporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (*fakeMetricExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(kind)
}

func (exporter *fakeMetricExporter) Export(context.Context, *metricdata.ResourceMetrics) error {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	exporter.exports++
	return exporter.exportErr
}

func (exporter *fakeMetricExporter) ForceFlush(context.Context) error {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	exporter.flushes++
	return exporter.flushErr
}

func (exporter *fakeMetricExporter) Shutdown(context.Context) error {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	exporter.shutdowns++
	return exporter.shutdownErr
}

func (exporter *fakeMetricExporter) snapshot() (int, int, int) {
	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	return exporter.exports, exporter.flushes, exporter.shutdowns
}

func secretFailure() error {
	return errors.New("provider failed with bearer secret-token")
}
