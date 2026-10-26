package workhttp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const maxBodyBytes = 1 << 20

func readDocument(request *http.Request, allowEmpty bool) (json.RawMessage, error) {
	if request == nil || request.Body == nil {
		if allowEmpty {
			return json.RawMessage(`{}`), nil
		}
		return nil, errRequestInvalid
	}
	media, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return nil, errMediaUnsupported
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, maxBodyBytes+1))
	if err != nil {
		return nil, errRequestInvalid
	}
	if len(body) > maxBodyBytes {
		return nil, errBodyTooLarge
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 && allowEmpty {
		return json.RawMessage(`{}`), nil
	}
	if len(body) < 2 || body[0] != '{' || !json.Valid(body) {
		return nil, errRequestInvalid
	}
	return json.RawMessage(body), nil
}

func pathIDs(request *http.Request, pattern string) (map[string]string, error) {
	result := map[string]string{}
	for _, key := range []string{"board_id", "column_id", "task_id", "comment_id",
		"label_id", "checklist_id", "event_id", "rule_id"} {
		if !strings.Contains(pattern, "{"+key+"}") {
			continue
		}
		value := request.PathValue(key)
		parsed, err := uuid.Parse(value)
		if err != nil || parsed == uuid.Nil {
			return nil, errRequestInvalid
		}
		result[key] = parsed.String()
	}
	return result, nil
}

func idempotencyKey(request *http.Request, spec routeSpec, access Access,
	ids map[string]string, body []byte,
) (string, error) {
	values := request.Header.Values("Idempotency-Key")
	if len(values) == 1 && validKey(values[0]) {
		return values[0], nil
	}
	if len(values) != 0 || !spec.compatibility {
		return "", errRequestInvalid
	}
	hash := sha256.New()
	for _, value := range []string{access.ConsumerID(), access.BuildID(), spec.operation,
		ids["board_id"], ids["task_id"], ids["event_id"], string(body)} {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(value))
	}
	return "v1." + hex.EncodeToString(hash.Sum(nil)), nil
}

func validKey(value string) bool {
	if len(value) < 16 || len(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for _, symbol := range value {
		if !(symbol >= 'a' && symbol <= 'z' || symbol >= 'A' && symbol <= 'Z' ||
			symbol >= '0' && symbol <= '9' || strings.ContainsRune("._:-", symbol)) {
			return false
		}
	}
	return true
}
