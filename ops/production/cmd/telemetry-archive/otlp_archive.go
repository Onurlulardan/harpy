package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const maximumRequestBytes = 8 << 20

type archiveHandler struct {
	directory string
	clock     func() time.Time
}

func (handler archiveHandler) ServeHTTP(
	response http.ResponseWriter,
	request *http.Request,
) {
	if request.Method == http.MethodGet && request.URL.Path == "/livez" {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	signal, validPath := map[string]string{
		"/v1/traces":  "traces",
		"/v1/metrics": "metrics",
		"/v1/logs":    "logs",
	}[request.URL.Path]
	if request.Method != http.MethodPost || !validPath ||
		request.URL.RawQuery != "" {
		http.NotFound(response, request)
		return
	}
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/x-protobuf" && contentType != "application/json" {
		response.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumRequestBytes)
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = handler.archive(signal, contentType, body); err != nil {
		response.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Content-Type", contentType)
	response.WriteHeader(http.StatusOK)
	if contentType == "application/json" {
		_, _ = response.Write([]byte("{}"))
	}
}

func (handler archiveHandler) archive(signal, contentType string, body []byte) error {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("generate archive identity")
	}
	now := handler.clock().UTC()
	extension := ".pb"
	if contentType == "application/json" {
		extension = ".json"
	}
	name := fmt.Sprintf("%s-%s-%s%s", now.Format("20060102T150405.000000000Z"),
		signal, hex.EncodeToString(nonce[:]), extension)
	path := filepath.Join(handler.directory, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create telemetry archive")
	}
	if _, err = file.Write(body); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(path)
		return fmt.Errorf("write telemetry archive")
	}
	return nil
}
