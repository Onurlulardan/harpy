package httpedge

import (
	"bytes"
	"fmt"
	"net/http"
)

type responseBuffer struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newResponseBuffer() *responseBuffer {
	return &responseBuffer{header: make(http.Header)}
}

func (buffer *responseBuffer) Header() http.Header {
	return buffer.header
}

func (buffer *responseBuffer) WriteHeader(status int) {
	if buffer.status == 0 {
		buffer.status = status
	}
}

func (buffer *responseBuffer) Write(content []byte) (int, error) {
	if buffer.status == 0 {
		buffer.status = http.StatusOK
	}
	return buffer.body.Write(content)
}

func (buffer *responseBuffer) statusCode() int {
	if buffer.status == 0 {
		return http.StatusOK
	}
	return buffer.status
}

func (buffer *responseBuffer) flush(target http.ResponseWriter) error {
	for key, values := range buffer.header {
		target.Header()[key] = append([]string(nil), values...)
	}
	status := buffer.status
	if status == 0 {
		status = http.StatusOK
	}
	target.WriteHeader(status)
	if _, err := target.Write(buffer.body.Bytes()); err != nil {
		return fmt.Errorf("flush buffered HTTP response: %w", err)
	}
	return nil
}
