package iam

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

const systemTransferJSONLimit int64 = 2 << 20

var errSystemTransferRequest = errors.New("IAM transfer HTTP request is invalid")
var systemTransferKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)

type systemExportSubmitDocument struct {
	Format string          `json:"format"`
	Query  json.RawMessage `json:"query"`
}

type systemImportSubmitDocument struct {
	SourceID        string          `json:"source_id"`
	DuplicateAction string          `json:"duplicate_action"`
	Mapping         json.RawMessage `json:"mapping"`
}

func decodeSystemTransferJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var document T
	if err := httpinput.DecodeJSON(w, r, systemTransferJSONLimit, &document); err != nil {
		return document, err
	}
	return document, nil
}

func systemTransferPath(r *http.Request) (SystemTransferResource, bool) {
	if r == nil {
		return "", false
	}
	resource, err := ParseSystemTransferResource(r.PathValue("resource"))
	return resource, err == nil
}

func systemTransferV7Path(r *http.Request, name string) (uuid.UUID, bool) {
	if r == nil {
		return uuid.Nil, false
	}
	wire := r.PathValue(name)
	value, err := uuid.Parse(wire)
	return value, err == nil && value.Version() == 7 && value.Variant() == uuid.RFC4122 &&
		value.String() == wire
}

func systemTransferIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil || len(r.Header.Values("Idempotency-Key")) != 1 {
		return "", false
	}
	value := r.Header.Get("Idempotency-Key")
	return value, systemTransferKeyPattern.MatchString(value)
}

func systemTransferImportFormat(r *http.Request) (string, error) {
	if r == nil || len(r.Header.Values("Content-Type")) != 1 ||
		len(r.Header.Values("Content-Encoding")) != 0 {
		return "", httpinput.ErrMediaTypeUnsupported
	}
	mediaType, parameters, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || len(parameters) > 1 {
		return "", httpinput.ErrMediaTypeUnsupported
	}
	if charset, exists := parameters["charset"]; exists &&
		!strings.EqualFold(charset, "utf-8") {
		return "", httpinput.ErrMediaTypeUnsupported
	}
	if mediaType == "text/csv" {
		return "csv", nil
	}
	if mediaType == "application/x-ndjson" {
		return "jsonl", nil
	}
	return "", httpinput.ErrMediaTypeUnsupported
}

func systemTransferReportPage(r *http.Request) (int64, int, bool) {
	if r == nil || r.URL == nil {
		return 0, 0, false
	}
	query := r.URL.Query()
	for name, values := range query {
		if (name != "after_row" && name != "limit") || len(values) != 1 {
			return 0, 0, false
		}
	}
	after, afterErr := strconv.ParseInt(query.Get("after_row"), 10, 64)
	if query.Get("after_row") == "" {
		after, afterErr = 0, nil
	}
	limit, limitErr := strconv.Atoi(query.Get("limit"))
	if query.Get("limit") == "" {
		limit, limitErr = 50, nil
	}
	return after, limit, afterErr == nil && limitErr == nil && after >= 0 &&
		after <= 1_000_000 && limit >= 1 && limit <= 100
}

func systemTransferQueryAbsent(r *http.Request) bool {
	return r != nil && r.URL != nil && r.URL.RawQuery == ""
}
