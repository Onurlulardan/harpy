package cataloghttp

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

const maximumBodyBytes = 1 << 20

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)

func DecodeBody[T any](request *http.Request, target *T) error {
	if request == nil || request.Body == nil || target == nil ||
		len(request.Header.Values("Content-Type")) != 1 {
		return ErrInvalid
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return ErrInvalid
	}
	limited := http.MaxBytesReader(nil, request.Body, maximumBodyBytes)
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(target); err != nil {
		return ErrInvalid
	}
	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrInvalid
	}
	return nil
}

func RequireNoBody(request *http.Request) error {
	if request == nil || request.Body == nil || request.ContentLength > 0 ||
		len(request.TransferEncoding) > 0 || len(request.Header.Values("Content-Type")) > 0 {
		return ErrInvalid
	}
	if request.Body == http.NoBody {
		return nil
	}
	contents, err := io.ReadAll(io.LimitReader(request.Body, 1))
	if err != nil || len(contents) != 0 {
		return ErrInvalid
	}
	return nil
}

func RequireNoQuery(request *http.Request) error {
	if request == nil || request.URL == nil || request.URL.RawQuery != "" {
		return ErrInvalid
	}
	return nil
}

func PageInput(request *http.Request) (int, *uuid.UUID, error) {
	return pageInput(request, wireV7)
}

func PackagePageInput(request *http.Request) (int, *uuid.UUID, error) {
	return pageInput(request, wirePackageID)
}

func pageInput(
	request *http.Request,
	parseIdentity func(string) (uuid.UUID, bool),
) (int, *uuid.UUID, error) {
	if request == nil || request.URL == nil {
		return 0, nil, ErrInvalid
	}
	query, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		return 0, nil, ErrInvalid
	}
	for key, values := range query {
		if (key != "limit" && key != "after") || len(values) != 1 {
			return 0, nil, ErrInvalid
		}
	}
	limit := 50
	if values, found := query["limit"]; found {
		parsed, parseErr := strconv.Atoi(values[0])
		if parseErr != nil || parsed < 1 || parsed > 100 || strconv.Itoa(parsed) != values[0] {
			return 0, nil, ErrInvalid
		}
		limit = parsed
	}
	var after *uuid.UUID
	if values, found := query["after"]; found {
		parsed, ok := parseIdentity(values[0])
		if !ok {
			return 0, nil, ErrInvalid
		}
		after = &parsed
	}
	return limit, after, nil
}

func PathV7(request *http.Request, name string) (uuid.UUID, error) {
	if request == nil {
		return uuid.Nil, ErrInvalid
	}
	value, ok := wireV7(request.PathValue(name))
	if !ok {
		return uuid.Nil, ErrInvalid
	}
	return value, nil
}

func PathPackageID(request *http.Request, name string) (uuid.UUID, error) {
	if request == nil {
		return uuid.Nil, ErrInvalid
	}
	value, err := ParsePackageID(request.PathValue(name))
	if err != nil {
		return uuid.Nil, ErrInvalid
	}
	return value, nil
}

func ParsePackageID(wire string) (uuid.UUID, error) {
	value, ok := wirePackageID(wire)
	if !ok {
		return uuid.Nil, ErrInvalid
	}
	return value, nil
}

func wireV7(wire string) (uuid.UUID, bool) {
	value, err := uuid.Parse(wire)
	return value, err == nil && value != uuid.Nil && value.Version() == 7 &&
		value.Variant() == uuid.RFC4122 && value.String() == wire
}

func wirePackageID(wire string) (uuid.UUID, bool) {
	identity, err := packagecontract.ParsePackageID(wire)
	if err != nil || identity.String() != wire {
		return uuid.Nil, false
	}
	value, err := uuid.Parse(wire)
	return value, err == nil && value != uuid.Nil && value.String() == wire
}

func PathVersion(request *http.Request) (string, error) {
	if request == nil {
		return "", ErrInvalid
	}
	wire := request.PathValue("version")
	value, err := packagecontract.ParseVersion(wire)
	if err != nil || value.String() != wire {
		return "", ErrInvalid
	}
	return wire, nil
}

func IdempotencyKey(request *http.Request) (string, error) {
	if request == nil {
		return "", ErrInvalid
	}
	values := request.Header.Values("Idempotency-Key")
	if len(values) != 1 || strings.TrimSpace(values[0]) != values[0] ||
		!idempotencyPattern.MatchString(values[0]) {
		return "", ErrInvalid
	}
	return values[0], nil
}
