package telemetry

import "testing"

func TestParseExportHeadersUsesBaggageEncoding(t *testing.T) {
	headers, err := parseExportHeaders("authorization=Bearer%20token,x-api-key=value")
	if err != nil {
		t.Fatal(err)
	}
	if headers["Authorization"] != "Bearer token" || headers["X-Api-Key"] != "value" {
		t.Fatalf("unexpected parsed headers: %#v", headers)
	}
}

func TestParseExportHeadersRejectsAmbiguousOrTransportHeaders(t *testing.T) {
	for _, raw := range []string{
		"x-key=one,X-Key=two",
		"authorization=value;metadata=yes",
		"content-length=10",
		"traceparent=value",
		"invalid member",
		"x-key=%zz",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := parseExportHeaders(raw); err == nil {
				t.Fatalf("invalid OTLP headers accepted: %s", raw)
			}
		})
	}
}
