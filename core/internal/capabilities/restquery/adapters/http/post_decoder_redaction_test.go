package httpquery

import (
	"fmt"
	"strings"
	"testing"
)

func TestPOSTDecoderDoesNotReflectRejectedValuesInErrors(t *testing.T) {
	t.Parallel()
	const secret = "never-log-this-value"
	_, err := NewPOSTDecoder().Decode([]byte(
		`{"where":{"field":"payload","op":"eq","value":"` + secret +
			`","unexpected":true}}`,
	))
	if err == nil {
		t.Fatal("Decode() error = nil")
	}
	rendered := fmt.Sprintf("%v|%+v|%#v", err, err, err)
	if strings.Contains(rendered, secret) || strings.Contains(rendered, "unexpected") {
		t.Fatalf("error reflected rejected input: %s", rendered)
	}
}

func TestPOSTDecoderRequestFormattingRedactsLiteralValues(t *testing.T) {
	t.Parallel()
	const secret = "request-secret-value"
	request := decodeSuccess(t, `{"where":{"field":"payload","op":"eq","value":"`+secret+`"}}`)
	rendered := fmt.Sprintf("%v|%+v|%#v", request, request, request)
	if strings.Contains(rendered, secret) || !strings.Contains(rendered, "[REDACTED]") {
		t.Fatalf("request formatting is not redacted: %s", rendered)
	}
}
