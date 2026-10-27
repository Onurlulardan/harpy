package httptransport

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)

func PathV7(request *http.Request, name string) (uuid.UUID, bool) {
	if request == nil {
		return uuid.Nil, false
	}
	wire := request.PathValue(name)
	value, err := uuid.Parse(wire)
	return value, err == nil && value != uuid.Nil && value.Version() == 7 &&
		value.Variant() == uuid.RFC4122 && value.String() == wire
}

func IdempotencyKey(request *http.Request) (string, bool) {
	if request == nil {
		return "", false
	}
	values := request.Header.Values("Idempotency-Key")
	if len(values) != 1 || !idempotencyPattern.MatchString(values[0]) {
		return "", false
	}
	return values[0], true
}

func PageInput(request *http.Request) (int, *uuid.UUID, bool) {
	if request == nil || request.URL == nil {
		return 0, nil, false
	}
	query, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		return 0, nil, false
	}
	for key, values := range query {
		if (key != "limit" && key != "after") || len(values) != 1 {
			return 0, nil, false
		}
	}
	limit := 50
	if values, ok := query["limit"]; ok {
		parsed, parseErr := strconv.Atoi(values[0])
		if parseErr != nil || parsed < 1 || parsed > 100 || strconv.Itoa(parsed) != values[0] {
			return 0, nil, false
		}
		limit = parsed
	}
	var after *uuid.UUID
	if values, ok := query["after"]; ok {
		parsed, parseOK := WireV7(values[0])
		if !parseOK {
			return 0, nil, false
		}
		after = &parsed
	}
	return limit, after, true
}

func WireV7(wire string) (uuid.UUID, bool) {
	value, err := uuid.Parse(wire)
	return value, err == nil && value.Version() == 7 && value.Variant() == uuid.RFC4122 &&
		value.String() == wire
}

func WireUUID(wire string) (uuid.UUID, bool) {
	value, err := uuid.Parse(wire)
	return value, err == nil && value != uuid.Nil && value.Variant() == uuid.RFC4122 &&
		value.String() == wire
}

func BodyAbsent(request *http.Request) bool {
	if request == nil || request.ContentLength > 0 || len(request.TransferEncoding) > 0 ||
		len(request.Header.Values("Content-Type")) > 0 || request.Body == nil {
		return false
	}
	if request.Body == http.NoBody {
		return true
	}
	contents, err := io.ReadAll(io.LimitReader(request.Body, 1))
	return err == nil && len(contents) == 0
}
