package apidocs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const maximumDocumentBytes = 8 << 20

type documentHeader struct {
	OpenAPI string `json:"openapi"`
	Info    struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	Paths map[string]json.RawMessage `json:"paths"`
}

func validateDocument(raw []byte) ([]byte, string, error) {
	if len(raw) == 0 || len(raw) > maximumDocumentBytes {
		return nil, "", fmt.Errorf("API documentation artifact size is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var header documentHeader
	if err := decoder.Decode(&header); err != nil {
		return nil, "", fmt.Errorf("decode API documentation artifact: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, "", fmt.Errorf("API documentation artifact has trailing content")
	}
	if header.OpenAPI != "3.1.1" || strings.TrimSpace(header.Info.Title) == "" ||
		strings.TrimSpace(header.Info.Version) == "" || header.Paths == nil {
		return nil, "", fmt.Errorf("API documentation artifact contract is invalid")
	}
	digest := sha256.Sum256(raw)
	return append([]byte(nil), raw...), `"` + hex.EncodeToString(digest[:]) + `"`, nil
}
