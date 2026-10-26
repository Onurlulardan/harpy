package documentshttp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCurrentDownloadIsRawPrivateBinary(t *testing.T) {
	body := []byte("private encrypted-at-rest plaintext response")
	digest := sha256.Sum256(body)
	handler := newTestHandler(t, applicationStub(func(_ context.Context,
		request Request,
	) (Response, error) {
		if request.Operation != operationDownload || request.ResourceID != testDocumentID {
			t.Fatalf("request = %#v", request)
		}
		return Response{Status: http.StatusOK, Binary: &Binary{
			Filename: "rapor\r\n2026.pdf", MIME: "application/pdf",
			Bytes: int64(len(body)), SHA256: digest,
			Content: ioCloser{Reader: bytes.NewReader(body)},
		}}, nil
	}))
	routes := routeMux(handler.Routes())
	request := httptest.NewRequest(http.MethodGet,
		"/api/documents/"+testDocumentID+"/download", nil)
	response := httptest.NewRecorder()
	routes.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), body) {
		t.Fatalf("response = %d %q", response.Code, response.Body.Bytes())
	}
	header := response.Header()
	if header.Get("Cache-Control") != "private, no-store" ||
		header.Get("X-Content-Type-Options") != "nosniff" ||
		header.Get("Content-Type") != "application/pdf" {
		t.Fatalf("binary headers = %#v", header)
	}
	disposition := header.Get("Content-Disposition")
	if !strings.HasPrefix(disposition, "attachment;") ||
		strings.ContainsAny(disposition, "\r\n") {
		t.Fatalf("unsafe disposition = %q", disposition)
	}
}

type ioCloser struct{ *bytes.Reader }

func (ioCloser) Close() error { return nil }
