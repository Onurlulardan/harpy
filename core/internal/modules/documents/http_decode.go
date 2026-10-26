package documents

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"time"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func decodeHTTPRequest[T any](document json.RawMessage) (T, error) {
	var result T
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if len(document) == 0 || decoder.Decode(&result) != nil {
		return result, documentshttp.ErrInvalid
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return result, documentshttp.ErrInvalid
	}
	return result, nil
}

func parseHTTPID(value string, optional bool) (foundation.ID, error) {
	if value == "" && optional {
		return foundation.ID{}, nil
	}
	identifier, err := foundation.ParseID(value)
	if err != nil || !identifier.UUIDv7() {
		return foundation.ID{}, documentshttp.ErrInvalid
	}
	return identifier, nil
}

func parseHTTPDigest(value string) ([32]byte, error) {
	var result [32]byte
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(result) || hex.EncodeToString(decoded) != value {
		return result, documentshttp.ErrInvalid
	}
	copy(result[:], decoded)
	return result, nil
}

func httpQueryValue(values url.Values, key string) (string, bool) {
	items, found := values[key]
	return func() (string, bool) {
		if !found || len(items) == 0 {
			return "", true
		}
		return items[0], len(items) == 1 && items[0] != ""
	}()
}

func httpLimit(values url.Values, maximum uint16) (uint16, error) {
	raw, valid := httpQueryValue(values, "limit")
	if !valid {
		return 0, documentshttp.ErrInvalid
	}
	if raw == "" {
		return 50, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 16)
	if err != nil || parsed < 1 || parsed > uint64(maximum) {
		return 0, documentshttp.ErrInvalid
	}
	return uint16(parsed), nil
}

func httpPositiveUint(values url.Values, key string) (uint64, error) {
	raw, valid := httpQueryValue(values, key)
	if !valid {
		return 0, documentshttp.ErrInvalid
	}
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, documentshttp.ErrInvalid
	}
	return value, nil
}

func httpTime(values url.Values, key string) (time.Time, error) {
	raw, valid := httpQueryValue(values, key)
	if !valid {
		return time.Time{}, documentshttp.ErrInvalid
	}
	if raw == "" {
		return time.Time{}, nil
	}
	value, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil || value.Location() != time.UTC {
		return time.Time{}, documentshttp.ErrInvalid
	}
	return value, nil
}

func onlyHTTPQuery(values url.Values, allowed ...string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		set[key] = struct{}{}
	}
	for key := range values {
		if _, found := set[key]; !found {
			return false
		}
	}
	return true
}
