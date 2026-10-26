package telemetry

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOTLPHTTPProtobufExportsBothSignals(t *testing.T) {
	receiver := &otlpReceiver{}
	server := httptest.NewServer(http.HandlerFunc(receiver.receive))
	defer server.Close()
	const headers = "authorization=Bearer%20smoke-token,x-api-key=header-value"
	settings := telemetrySettingsForEndpoint(t, server.URL, headers)
	output := &lockedBuffer{}
	runtime, err := New(
		context.Background(), settings, telemetryBuildInfo(t), testLogger(output),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, span := runtime.TracerProvider().Tracer("test/otlp").Start(context.Background(), "smoke")
	span.End()
	counter, err := runtime.MeterProvider().Meter("test/otlp").Int64Counter("smoke.operations")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(context.Background(), 1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := runtime.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	requests, receiverErr := receiver.snapshot()
	if receiverErr != "" {
		t.Fatal(receiverErr)
	}
	if requests["/v1/traces"] == 0 || requests["/v1/metrics"] == 0 {
		t.Fatalf("OTLP signals missing: %#v", requests)
	}
	if strings.Contains(output.String(), "smoke-token") || strings.Contains(output.String(), "header-value") {
		t.Fatalf("OTLP credential leaked to log: %s", output.String())
	}
}

type otlpReceiver struct {
	mu       sync.Mutex
	requests map[string]int
	err      string
}

func (receiver *otlpReceiver) receive(writer http.ResponseWriter, request *http.Request) {
	problem := validateOTLPRequest(request)
	receiver.mu.Lock()
	if receiver.requests == nil {
		receiver.requests = make(map[string]int)
	}
	receiver.requests[request.URL.Path]++
	if receiver.err == "" {
		receiver.err = problem
	}
	receiver.mu.Unlock()
	writer.Header().Set("Content-Type", "application/x-protobuf")
	writer.WriteHeader(http.StatusOK)
}

func (receiver *otlpReceiver) snapshot() (map[string]int, string) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	requests := make(map[string]int, len(receiver.requests))
	for path, count := range receiver.requests {
		requests[path] = count
	}
	return requests, receiver.err
}

func validateOTLPRequest(request *http.Request) string {
	if request.Method != http.MethodPost {
		return "OTLP request method is not POST"
	}
	if request.Header.Get("Content-Type") != "application/x-protobuf" {
		return "OTLP request content type is not protobuf"
	}
	if request.Header.Get("Content-Encoding") != "gzip" {
		return "OTLP request compression is not gzip"
	}
	if request.Header.Get("Authorization") != "Bearer smoke-token" ||
		request.Header.Get("X-Api-Key") != "header-value" {
		return "OTLP configured headers are missing"
	}
	reader, err := gzip.NewReader(request.Body)
	if err != nil {
		return "OTLP gzip body is invalid"
	}
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil || len(body) == 0 {
		return "OTLP protobuf body is empty"
	}
	return ""
}
