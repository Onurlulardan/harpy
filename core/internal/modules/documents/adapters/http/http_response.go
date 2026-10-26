package documentshttp

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *Handler) writeResponse(w http.ResponseWriter, r *http.Request,
	response Response,
) {
	if response.Binary != nil {
		handler.writeBinary(w, r, response)
		return
	}
	status := response.Status
	if status < 200 || status > 299 {
		handler.writeFailure(w, r, ErrUnavailable)
		return
	}
	if response.Location != "" {
		w.Header().Set("Location", response.Location)
	}
	handler.writeJSON(w, r, status, successEnvelope{Success: true, Data: response.Data})
}

func (handler *Handler) writeJSON(w http.ResponseWriter, r *http.Request,
	status int, payload any,
) {
	body, err := json.Marshal(payload)
	if err != nil {
		handler.report(errors.New("Documents HTTP response marshal failed"))
		return
	}
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", handler.correlation(r))
	header.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if _, err = w.Write(body); err != nil {
		handler.report(errors.New("Documents HTTP response write failed"))
	}
}

func (handler *Handler) writeBinary(w http.ResponseWriter, r *http.Request,
	response Response,
) {
	binary := response.Binary
	if response.Status != http.StatusOK || !validBinary(binary) {
		if binary != nil && binary.Content != nil {
			_ = binary.Content.Close()
		}
		handler.writeFailure(w, r, ErrUnavailable)
		return
	}
	defer func() {
		if err := binary.Content.Close(); err != nil {
			handler.report(errors.New("Documents binary close failed"))
		}
	}()
	header := w.Header()
	header.Set("Content-Type", safeMediaType(binary.MIME))
	header.Set("Content-Disposition", attachment(binary.Filename))
	header.Set("Content-Length", strconv.FormatInt(binary.Bytes, 10))
	header.Set("Cache-Control", "private, no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Content-SHA256", hex.EncodeToString(binary.SHA256[:]))
	header.Set("X-Correlation-ID", handler.correlation(r))
	w.WriteHeader(http.StatusOK)
	written, err := io.CopyN(w, binary.Content, binary.Bytes)
	if err != nil || written != binary.Bytes {
		handler.report(errors.New("Documents binary response truncated"))
	}
}

func validBinary(value *Binary) bool {
	return value != nil && value.Content != nil && value.Bytes > 0 &&
		value.SHA256 != [32]byte{} && safeFilename(value.Filename) != ""
}

func attachment(filename string) string {
	return mime.FormatMediaType("attachment", map[string]string{
		"filename": safeFilename(filename),
	})
}

func safeFilename(value string) string {
	value = strings.Map(func(character rune) rune {
		if character < 0x20 || character == 0x7f || character == '/' || character == '\\' {
			return -1
		}
		return character
	}, value)
	value = strings.TrimSpace(value)
	if len([]byte(value)) > 255 {
		return "download"
	}
	return value
}

func safeMediaType(value string) string {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil || !strings.Contains(mediaType, "/") {
		return "application/octet-stream"
	}
	return fmt.Sprintf("%s", mediaType)
}
