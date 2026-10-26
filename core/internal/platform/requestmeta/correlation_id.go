package requestmeta

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
)

const CorrelationHeader = "X-Correlation-ID"

const EntropyFailureCorrelationID CorrelationID = "correlation-id-unavailable"

type CorrelationID string

func ResolveCorrelationID(header http.Header) (CorrelationID, error) {
	return resolveCorrelationID(header, randomid.NewUUIDv4)
}

func resolveCorrelationID(header http.Header, generate func() (string, error)) (CorrelationID, error) {
	values := header.Values(CorrelationHeader)
	if len(values) == 1 && ValidCorrelationID(values[0]) {
		return CorrelationID(values[0]), nil
	}
	value, err := generate()
	return CorrelationID(value), err
}

func ValidCorrelationID(value string) bool {
	if len(value) == 0 || len(value) > 64 || !asciiAlphaNumeric(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if !asciiAlphaNumeric(character) && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func asciiAlphaNumeric(character byte) bool {
	return character >= 'A' && character <= 'Z' ||
		character >= 'a' && character <= 'z' ||
		character >= '0' && character <= '9'
}

func (id CorrelationID) String() string {
	return string(id)
}
